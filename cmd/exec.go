package cmd

import (
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
		createExecutor()
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execProvisionCmd.Flags().StringVarP(&execName, "name", "n", "", "Name of spec to provision")
	execProvisionCmd.Flags().StringVarP(&execNamespace, "namespace", "p", "", "Namespace to provision bundle to")
	execCmd.AddCommand(execProvisionCmd)
}

func createExecutor() {
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
			fmt.Printf("HERE!")
			targetSpec = s
		}
	}
	if targetSpec == nil {
		fmt.Printf("Didn't find supplied APB")
		return
	}
	labels := map[string]string{
		"bundle-fqname":   targetSpec.FQName,
		"bundle-action":   "provision",
		"bundle-pod-name": pn,
	}
	ec := runtime.ExecutionContext{
		BundleName: pn,
		Targets:    targets,
		Metadata:   labels,
		Action:     "provision",
		Image:      targetSpec.Image,
		Account:    "default",
		Location:   execNamespace,
	}
	//	conf := runtime.Configuration{}
	//	runtime.NewRuntime(conf)
	k8scli, err := clients.Kubernetes()
	if err != nil {
		panic(err.Error())
	}

	pods, err := k8scli.Client.CoreV1().Pods(execNamespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting list of pods: %v", err)
		return
	}
	fmt.Printf("%v\n", len(pods.Items))
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
	fmt.Printf("HERE!!: %v", ec)
	_, err = k8scli.Client.CoreV1().Pods(execNamespace).Create(pod)
	if err != nil {
		fmt.Printf("Failed to create pod: %v", err)
	}
	return
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
