//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/automationbroker/apb/pkg/config"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/lestrrat/go-jsschema/validator"
	"github.com/pborman/uuid"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/api/core/v1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunBundle will run the bundle's action in the given namespace
func RunBundle(action string, ns string, bundleName string, sandboxRole string, bundleRegistry string, printLogs bool, skipParams bool, args []string) (podName string, err error) {
	reg := []config.Registry{}
	var id string
	var targetSpec *bundle.Spec
	var candidateSpecs []*bundle.Spec

	if action == "deprovision" {
		id, err = getProvisionedInstanceId(bundleName)
		if err != nil {
			return "", err
		}
	} else {
		id = uuid.New()
	}

	podName = fmt.Sprintf("bundle-%s-%s", action, id)
	config.Registries.UnmarshalKey("Registries", &reg)
	for _, r := range reg {
		if len(bundleRegistry) > 0 && r.Config.Name != bundleRegistry {
			continue
		}
		for _, s := range r.Specs {
			if s.FQName == bundleName {
				candidateSpecs = append(candidateSpecs, s)
				fmt.Printf("Found APB [%v] in registry [%v]\n", bundleName, r.Config.Name)
			}
		}
	}
	if len(candidateSpecs) == 0 {
		if len(bundleRegistry) > 0 {
			return "", errors.New(fmt.Sprintf("failed to find APB [%v] in registry [%v]", bundleName, bundleRegistry))
		}
		return "", errors.New(fmt.Sprintf("failed to find APB [%v] in configured registries", bundleName))
		// TODO: return an ErrorBundleNotFound
	}
	if len(candidateSpecs) > 1 {
		return "", errors.New(fmt.Sprintf("found multiple APBs with matching name [%v]. Specify a registry with --registry", bundleName))
	}

	targetSpec = candidateSpecs[0]
	log.Debugf("na: %v", targetSpec.FQName)

	// determine the correct plan
	plan := selectPlan(targetSpec)
	if plan.Name == "" {
		log.Warning("Did not find a selected plan")
	} else {
		fmt.Printf("Plan: %v\n", plan.Name)
	}

	var params bundle.Parameters
	if skipParams {
		params = bundle.Parameters{}
	} else {
		params, err = selectParameters(plan)
		if err != nil {
			return "", err
		}
	}

	extraVars, err := createExtraVars(id, ns, &params, plan)
	if err != nil {
		return "", err
	}

	labels := map[string]string{
		"bundle-fqname":   targetSpec.FQName,
		"bundle-action":   action,
		"bundle-pod-name": podName,
	}

	// TODO: using edit directly. The bundle code uses clusterConfig.SandboxRole
	// which is defined by the template. So far we've been using edit.

	runtime.NewRuntime(runtime.Configuration{})
	targets := []string{ns}
	serviceAccount, namespace, err := runtime.Provider.CreateSandbox(podName, ns, targets, sandboxRole, labels)
	if err != nil {
		fmt.Printf("\nProblem creating sandbox [%s] to run APB. Did you run `oc new-project %s` first?\n\n", podName, ns)
		log.Errorf("error creating sandbox: %v", err)
		os.Exit(-1)
	}

	ec := runtime.ExecutionContext{
		BundleName: podName,
		Targets:    targets,
		Metadata:   labels,
		Action:     action,
		Image:      targetSpec.Image,
		Account:    serviceAccount,
		Location:   namespace,
		ExtraVars:  extraVars,
	}

	k8scli, err := clients.Kubernetes()
	if err != nil {
		// TODO: return err
		panic(err.Error())
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ec.BundleName,
			Labels: ec.Metadata,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  podName,
					Image: ec.Image,
					Args: []string{
						ec.Action,
						"--extra-vars",
						ec.ExtraVars,
					},
					Env:             createPodEnv(ec),
					ImagePullPolicy: "Always",
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: ec.Account,
		},
	}
	_, err = k8scli.Client.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		return "", err
	}
	fmt.Printf("Successfully created pod [%v] to %s [%v] in namespace [%v]\n", podName, ec.Action, bundleName, ns)

	if printLogs {
		printBundleLogs(podName, ns, action)
	}

	return
}

func GetPodStatus(namespace string, podName string) (string, error) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		panic(err.Error())
	}
	pod, err := k8scli.Client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	status := pod.Status.Phase
	return string(status), nil
}

func printBundleLogs(podName string, namespace string, action string) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		panic(err.Error())
	}

	logTailRequest := k8scli.Client.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{Follow: true})

	var requestStream io.ReadCloser
	var podStarted bool

	for podStarted == false {
		requestStream, err = logTailRequest.Stream()
		if err != nil {
			fmt.Printf("Waiting for APB %v pod [%v] to start...\n", action, podName)
			log.Debugf("%v", err)
			time.Sleep(3 * time.Second)
		} else {
			fmt.Printf("Pod started. Reading logs...\n")
			podStarted = true
		}
	}
	defer requestStream.Close()

	fmt.Println("-+- ---------------------- -+-")
	fmt.Println(" |         APB LOGS         | ")
	fmt.Println("-+- ---------------------- -+-")

	buf := make([]byte, 100)
	var doneReading bool
	for doneReading == false {
		n, err := requestStream.Read(buf)
		if err == io.EOF {
			doneReading = true
		}
		fmt.Printf("%s", buf[:n])
	}
}

func selectPlan(spec *bundle.Spec) bundle.Plan {
	var planName string
	var check = true
	for check {
		if len(spec.Plans) > 1 {
			fmt.Printf("List of available plans:\n")
			for _, plan := range spec.Plans {
				fmt.Printf("name: %v\n", plan.Name)
			}
		} else {
			return spec.Plans[0]
		}
		fmt.Printf("Enter name of plan to execute: ")
		fmt.Scanln(&planName)
		for _, plan := range spec.Plans {
			if plan.Name == planName {
				return plan
			}
		}
		fmt.Printf("Did not find plan [%v], try again.\n\n", planName)
	}
	return bundle.Plan{}
}

func selectParameters(plan bundle.Plan) (bundle.Parameters, error) {
	schemaPlan, err := bundle.ConvertPlansToSchema([]bundle.Plan{plan})
	if err != nil {
		log.Errorf("Error converting APB plans to JSON Schema: %v", err)
		return nil, err
	}
	planSchema := schemaPlan[0].Schemas
	schemaParams := planSchema.ServiceInstance.Create["parameters"]
	params := bundle.Parameters{}
	for _, param := range plan.Parameters {
		var inputValid = false
		var paramDefault interface{}

		if param.Default != nil {
			paramDefault = param.Default
		}

		for !inputValid {
			var paramInput string

			if len(param.Description) > 0 {
				fmt.Printf("Enter value for parameter [%v] (%v), default: [%v]: ", param.Name, param.Description, paramDefault)
			} else {
				fmt.Printf("Enter value for parameter [%v], default: [%v]: ", param.Name, paramDefault)
			}

			if param.DisplayType == "password" {
				passwordInputBytes, err := terminal.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err != nil {
					log.Errorf("Error while collecting password: %v", err)
					continue
				}
				paramInput = string(passwordInputBytes)
			} else {
				fmt.Scanln(&paramInput)
			}

			if paramInput == "" {
				switch paramDefault.(type) {
				case int:
					paramInput = strconv.Itoa(paramDefault.(int))
				case string:
					paramInput = paramDefault.(string)
				case float64:
					paramInput = strconv.FormatFloat(paramDefault.(float64), 'f', 0, 32)
				case bool:
					paramInput = strconv.FormatBool(paramDefault.(bool))
				}
			}
			if param.Required == true && paramInput == "" {
				fmt.Printf("Parameter [%v] is required. Please try again.\n", param.Name)
				continue
			}

			if len(param.Enum) > 0 {
				if !contains(param.Enum, paramInput) {
					fmt.Printf("[%v] is not a valid option. Available options: %v\n", paramInput, param.Enum)
					continue
				}
			}

			input, err := pruneInput(paramInput, param)
			if err != nil {
				fmt.Printf("Error accepting input: %v\n", err)
				fmt.Println("Please try again")
			} else {
				inputValid = true
				params.Add(param.Name, input)
			}
		}
	}
	v := validator.New(schemaParams)
	if err := v.Validate(params); err != nil {
		log.Debugf("Error validating parameters: %v", err)
		return nil, err
	}

	log.Debugf("Params: %v\n", params)
	return params, nil
}

func createPodEnv(executionContext runtime.ExecutionContext) []v1.EnvVar {
	podEnv := []v1.EnvVar{
		v1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		v1.EnvVar{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}
	return podEnv
}

func createExtraVars(id string, targetNamespace string, parameters *bundle.Parameters, plan bundle.Plan) (string, error) {
	var paramsCopy bundle.Parameters
	if parameters != nil && *parameters != nil {
		paramsCopy = *parameters
	} else {
		paramsCopy = make(bundle.Parameters)
	}

	if targetNamespace != "" {
		paramsCopy["namespace"] = targetNamespace
	}

	paramsCopy["cluster"] = "openshift"
	paramsCopy["_apb_plan_id"] = plan.Name
	paramsCopy["_apb_service_instance_id"] = id
	paramsCopy["_apb_service_class_id"] = id
	extraVars, err := json.Marshal(paramsCopy)
	return string(extraVars), err
}

func pruneInput(input string, param bundle.ParameterDescriptor) (interface{}, error) {
	var output interface{}
	var err error
	switch param.Type {
	case "string":
		output = input
	case "enum":
		output = input
	case "boolean", "bool":
		output, err = strconv.ParseBool(input)
		if err != nil {
			return nil, errors.New("Input must be a boolean")
		}
	case "integer", "int":
		output, err = strconv.ParseInt(input, 0, 0)
		if err != nil {
			return nil, errors.New("Input must be an integer")
		}
	case "number":
		output, err = strconv.ParseFloat(input, 64)
		if err != nil {
			return nil, errors.New("Input must be a float")
		}
	default:
		output = input
	}
	return output, nil
}

func contains(s []string, t string) bool {
	for _, str := range s {
		if str == t {
			return true
		}
	}
	return false
}

func getProvisionedInstanceId(name string) (string, error) {
	var instanceConfigs []config.ProvisionedInstance
	err := config.ProvisionedInstances.UnmarshalKey("ProvisionedInstances", &instanceConfigs)
	if err != nil {
		return "", err
	}
	for _, instance := range instanceConfigs {
		if instance.BundleName == name {
			if len(instance.InstanceIDs) == 1 {
				return instance.InstanceIDs[0], nil
			} else if len(instance.InstanceIDs) == 0 {
				return "", errors.New("found no available instances")
			} else {
				// Select instance
				fmt.Printf("Found more than one service instance for bundle [%v]:\n", name)
				for i, instance := range instance.InstanceIDs {
					fmt.Printf("[%v] - %v\n", i, instance)
				}
				var inputValid = false
				for !inputValid {
					var input string
					fmt.Printf("Enter the number of the instance ID you would wish to deprovision: ")
					fmt.Scanln(&input)
					if input == "" {
						continue
					}
					intInput, err := strconv.Atoi(input)
					if err != nil {
						fmt.Printf("Input was not a valid integer, please enter again.\n")
						continue
					}
					if intInput >= len(instance.InstanceIDs) || intInput < 0 {
						fmt.Printf("Input is out of range. Please select an integer from 0-%v\n", len(instance.InstanceIDs)-1)
						continue
					}
					return instance.InstanceIDs[intInput], nil
				}
			}
		}
	}
	return "", fmt.Errorf("No provisioned instances for bundle [%v]", name)
}
