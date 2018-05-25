package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/pborman/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	log "github.com/sirupsen/logrus"
)

var execName string
var execNamespace string

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Perform specified action on Service Bundles",
	Long:  `Perform actions (Provision, Deprovision, Bind, Unbind) on Service Bundles`,
}

var execProvisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision ServiceBundle images",
	Long:  `Provision ServiceBundles from a registry adapter`,
	Run: func(cmd *cobra.Command, args []string) {
		runBundle("provision")
	},
}

var execDeprovisionCmd = &cobra.Command{
	Use:   "deprovision",
	Short: "Deprovision ServiceBundle images",
	Long:  `Deprovision ServiceBundles from a registry adapter`,
	Run: func(cmd *cobra.Command, args []string) {
		runBundle("deprovision")
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execProvisionCmd.Flags().StringVarP(&execName, "name", "n", "", "Name of spec to provision")
	execProvisionCmd.Flags().StringVarP(&execNamespace, "namespace", "p", "", "Namespace to provision bundle to")
	execDeprovisionCmd.Flags().StringVarP(&execName, "name", "n", "", "Name of spec to provision")
	execDeprovisionCmd.Flags().StringVarP(&execNamespace, "namespace", "p", "", "Namespace to provision bundle to")

	execCmd.AddCommand(execProvisionCmd)
	execCmd.AddCommand(execDeprovisionCmd)
}

func runBundle(action string) {
	if execName == "" {
		fmt.Println("Failed to find --name argument. No bundle selected")
		return
	}
	if execNamespace == "" {
		fmt.Println("Failed to find --namespace argument. No namespace selected")
		return
	}
	specs := []*bundle.Spec{}
	var targetSpec *bundle.Spec
	targets := []string{execNamespace}
	pn := fmt.Sprintf("bundle-%s", uuid.New())
	viper.UnmarshalKey("Specs", &specs)
	for _, s := range specs {
		if s.FQName == execName {
			targetSpec = s
		}
	}
	if targetSpec == nil {
		fmt.Printf("Didn't find supplied APB: %v\n", execName)
		return
	}
	plan := selectPlan(targetSpec)
	if plan.Name == "" {
		fmt.Println("Did not find a selected plan")
	} else {
		fmt.Printf("Plan: %v\n", plan.Name)
	}
	params := selectParameters(plan)
	extraVars, err := createExtraVars(execNamespace, &params, plan)
	if err != nil {
		fmt.Printf("Error creating extravars: %v\n", err)
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
		Location:   execNamespace,
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
	_, err = k8scli.Client.CoreV1().Pods(execNamespace).Create(pod)
	if err != nil {
		fmt.Printf("Failed to create pod: %v", err)
	}
	fmt.Printf("Successfully created pod [%v] to %s [%v] in namespace [%v]\n", pn, ec.Action, execName, execNamespace)
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
		fmt.Printf("Enter value for parameter [%v], default: [%v]: ", param.Name, paramDefault)
		fmt.Scanln(&paramInput)
		if paramInput == "" {
			paramInput = paramDefault
		}
		params.Add(param.Name, paramInput)
	}
	fmt.Printf("Params: %v\n", params)
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
