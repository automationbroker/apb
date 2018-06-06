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

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/automationbroker/sbcli/util"
	"github.com/pborman/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Refresh - indicates whether we should refresh the list of images
var Refresh bool

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Interact with ServiceBundles",
	Long:  `List and execute ServiceBundles from configured registry adapters`,
}

var bundleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ServiceBundle images",
	Long:  `List ServiceBundles from a registry adapter`,
	Run: func(cmd *cobra.Command, args []string) {
		listImages()
	},
}

var bundleNamespace string

var bundleProvisionCmd = &cobra.Command{
	Use:   "provision <bundle name>",
	Short: "Provision ServiceBundle images",
	Long:  `Provision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runBundle("provision", args)
	},
}

var bundleDeprovisionCmd = &cobra.Command{
	Use:   "deprovision <bundle name>",
	Short: "Deprovision ServiceBundle images",
	Long:  `Deprovision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runBundle("deprovision", args)
	},
}

func init() {
	rootCmd.AddCommand(bundleCmd)

	bundleListCmd.Flags().BoolVarP(&Refresh, "refresh", "r", false, "refresh list of specs")
	bundleCmd.AddCommand(bundleListCmd)

	bundleProvisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "p", "", "Namespace to provision bundle to")
	bundleProvisionCmd.MarkFlagRequired("namespace")
	bundleCmd.AddCommand(bundleProvisionCmd)

	bundleDeprovisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "p", "", "Namespace to provision bundle to")
	bundleDeprovisionCmd.MarkFlagRequired("namespace")
	bundleCmd.AddCommand(bundleDeprovisionCmd)
}

func updateCachedList(specs []*bundle.Spec) error {
	viper.Set("Specs", specs)
	viper.WriteConfig()
	return nil
}

func getImages() ([]*bundle.Spec, error) {
	var regConfigList []registries.Config
	var regList []registries.Registry
	var specList []*bundle.Spec
	err := viper.UnmarshalKey("Registries", &regConfigList)
	if err != nil {
		log.Error("Error unmarshalling config: ", err)
		return nil, err
	}

	authNamespace := ""
	for _, config := range regConfigList {
		registry, err := registries.NewRegistry(config, authNamespace)
		if err != nil {
			log.Error("Error from creating a NewRegistry")
			log.Error(err)
			return nil, err
		}
		regList = append(regList, registry)
	}
	for _, reg := range regList {
		specs, count, err := reg.LoadSpecs()
		if err != nil {
			log.Errorf("registry: %v was unable to complete bootstrap - %v",
				reg.RegistryName(), err)
			return nil, err
		}
		log.Infof("Registry %v has %d bundles available from %d images scanned", reg.RegistryName(), len(specs), count)
		specList = append(specList, specs...)
	}

	return specList, nil
}

func printSpecs(specs []*bundle.Spec) {
	colFQName := util.TableColumn{Header: "BUNDLE"}
	colImage := util.TableColumn{Header: "IMAGE"}

	for _, s := range specs {
		colFQName.Data = append(colFQName.Data, s.FQName)
		colImage.Data = append(colImage.Data, s.Image)
	}

	tableToPrint := []util.TableColumn{colFQName, colImage}
	util.PrintTable(tableToPrint)
}

func listImages() {
	var specs []*bundle.Spec
	err := viper.UnmarshalKey("Specs", &specs)
	if err != nil {
		log.Error("Error unmarshalling config: ", err)
		return
	}
	if len(specs) > 0 && Refresh == false {
		fmt.Println("Found specs already in config")
		printSpecs(specs)
		return
	}
	specs, err = getImages()
	if err != nil {
		log.Error("Error getting images")
		return
	}
	err = updateCachedList(specs)
	if err != nil {
		log.Error("Error updating cache")
		return
	}
	printSpecs(specs)
}

func runBundle(action string, args []string) {
	bundleName := args[0]
	specs := []*bundle.Spec{}
	var targetSpec *bundle.Spec
	targets := []string{bundleNamespace}
	pn := fmt.Sprintf("bundle-%s", uuid.New())
	viper.UnmarshalKey("Specs", &specs)
	for _, s := range specs {
		if s.FQName == bundleName {
			targetSpec = s
		}
	}
	if targetSpec == nil {
		log.Errorf("Didn't find supplied APB: %v\n", bundleName)
		return
	}
	plan := selectPlan(targetSpec)
	if plan.Name == "" {
		log.Warning("Did not find a selected plan")
	} else {
		fmt.Printf("Plan: %v\n", plan.Name)
	}
	params := selectParameters(plan)
	extraVars, err := createExtraVars(bundleNamespace, &params, plan)
	if err != nil {
		log.Errorf("Error creating extravars: %v\n", err)
		return
	}

	labels := map[string]string{
		"bundle-fqname":   targetSpec.FQName,
		"bundle-action":   action,
		"bundle-pod-name": pn,
	}
	ec := runtime.ExecutionContext{
		BundleName: pn,
		Targets:    targets,
		Metadata:   labels,
		Action:     action,
		Image:      targetSpec.Image,
		Account:    "apb",
		Location:   bundleNamespace,
		ExtraVars:  extraVars,
	}
	//	conf := runtime.Configuration{}
	//	runtime.NewRuntime(conf)
	k8scli, err := clients.Kubernetes()
	if err != nil {
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
	_, err = k8scli.Client.CoreV1().Pods(bundleNamespace).Create(pod)
	if err != nil {
		log.Errorf("Failed to create pod: %v", err)
		return
	}
	fmt.Printf("Successfully created pod [%v] to %s [%v] in namespace [%v]\n", pn, ec.Action, bundleName, bundleNamespace)
	return
}

func selectPlan(spec *bundle.Spec) bundle.Plan {
	var planName string
	if len(spec.Plans) > 1 {
		fmt.Printf("List of available plans:\n")
		for _, plan := range spec.Plans {
			fmt.Printf("name: %v\n", plan.Name)
		}
		fmt.Printf("Enter name of plan you'd like to deploy: ")
		fmt.Scanln(&planName)
	} else {
		return spec.Plans[0]
	}
	for _, plan := range spec.Plans {
		if plan.Name == planName {
			return plan
		}
	}
	return bundle.Plan{}
}

func selectParameters(plan bundle.Plan) bundle.Parameters {
	params := bundle.Parameters{}
	for _, param := range plan.Parameters {
		var paramDefault string
		var paramInput string
		if param.Default != nil {
			paramDefault = param.Default.(string)
		}
		check := 1
		for check < 2 {
			fmt.Printf("Enter value for parameter [%v], default: [%v]: ", param.Name, paramDefault)
			fmt.Scanln(&paramInput)
			if paramInput == "" {
				paramInput = paramDefault
			}
			if param.Required == true && paramInput == "" {
				fmt.Printf("Parameter [%v] is required. Please try again.\n", param.Name)
			} else {
				check = 2
			}
		}
		params.Add(param.Name, paramInput)
	}
	log.Debugf("Params: %v\n", params)
	return params
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
