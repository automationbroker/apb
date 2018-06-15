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
	"os"
	"strconv"
	"syscall"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/lestrrat/go-jsschema/validator"
	"github.com/pborman/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/api/core/v1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunBundle will run the bundle's action in the given namespace
func RunBundle(action string, ns string, bundleName string, sandboxRole string, args []string) {
	reg := []Registry{}
	var targetSpec *bundle.Spec
	pn := fmt.Sprintf("bundle-%s", uuid.New())
	viper.UnmarshalKey("Registries", &reg)
	for _, r := range reg {
		for _, s := range r.Specs {
			if s.FQName == bundleName {
				targetSpec = s
			}
		}
	}
	if targetSpec == nil {
		log.Errorf("Didn't find supplied APB: %v\n", bundleName)

		// TODO: return an ErrorBundleNotFound
		return
	}

	// determine the correct plan
	plan := selectPlan(targetSpec)
	if plan.Name == "" {
		log.Warning("Did not find a selected plan")
	} else {
		fmt.Printf("Plan: %v\n", plan.Name)
	}

	params, err := selectParameters(plan)
	if err != nil {
		log.Errorf("Error validating selected parameters: %v", err)
		return
	}
	extraVars, err := createExtraVars(ns, &params, plan)
	if err != nil {
		log.Errorf("Error creating extravars: %v\n", err)
		return
	}

	labels := map[string]string{
		"bundle-fqname":   targetSpec.FQName,
		"bundle-action":   action,
		"bundle-pod-name": pn,
	}

	// TODO: using edit directly. The bundle code uses clusterConfig.SandboxRole
	// which is defined by the template. So far we've been using edit.

	runtime.NewRuntime(runtime.Configuration{})
	targets := []string{ns}
	serviceAccount, namespace, err := runtime.Provider.CreateSandbox(pn, ns, targets, sandboxRole, labels)
	if err != nil {
		fmt.Printf("\nProblem creating sandbox [%s] to run bundle. Did you run `oc new-project %s` first?\n\n", pn, ns)
		os.Exit(-1)
	}

	ec := runtime.ExecutionContext{
		BundleName: pn,
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
					Name:  pn,
					Image: ec.Image,
					Args: []string{
						ec.Action,
						"--extra-vars",
						ec.ExtraVars,
					},
					Env:             createPodEnv(ec),
					ImagePullPolicy: "IfNotPresent",
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: ec.Account,
		},
	}
	_, err = k8scli.Client.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		log.Errorf("Failed to create pod: %v", err)
		// TODO: return ErrorPodCreationFailed
		return
	}
	fmt.Printf("Successfully created pod [%v] to %s [%v] in namespace [%v]\n", pn, ec.Action, bundleName, ns)
	// TODO: return nil
	return
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
		fmt.Printf("Enter name of plan you'd like to deploy: ")
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
		log.Errorf("Error converting bundle plans to JSON Schema: %v", err)
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

func createExtraVars(targetNamespace string, parameters *bundle.Parameters, plan bundle.Plan) (string, error) {
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
	paramsCopy["_apb_service_instance_id"] = "1234"
	paramsCopy["_apb_service_class_id"] = "1234"
	paramsCopy["in_cluster"] = false
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
	case "bool":
		output, err = strconv.ParseBool(input)
		if err != nil {
			return nil, errors.New("Input must be a boolean")
		}
	case "int":
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
