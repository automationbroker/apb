package cmd

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	//"github.com/automationbroker/bundle-lib/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//	log "github.com/sirupsen/logrus"
)

var execName string

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
	execCmd.AddCommand(execProvisionCmd)
}

func createExecutor() {
	specs := []*bundle.Spec{}
	viper.UnmarshalKey("Specs", &specs)
	for _, s := range specs {
		if s.FQName == execName {
			ec := runtime.ExecutionContext{
				BundleName: s.FQName,
				//				Targets:
			}
			si.Spec = s
		}
	}
	fmt.Printf("Spec: %v", si.Spec)
	//	conf := runtime.Configuration{}
	//	runtime.NewRuntime(conf)
	k8scli, err := clients.Kubernetes()
	if err != nil {
		panic(err.Error())
	}
	ec := runtime.ExecutionContext{}
	ec.Name = si.Spec.FQName
	ec.Image = si.Spec.Image
	ec.Action = "provision"
	ec.Location = "dylan"
	pods, err := k8scli.Client.CoreV1().Pods("").List(metav1.ListOptions{})
	fmt.Printf("%v\n", len(pods.Items))
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   extContext.BundleName,
			Labels: extContext.Metadata,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  BundleContainerName,
					Image: extContext.Image,
					Args: []string{
						extContext.Action,
						"--extra-vars",
						extContext.ExtraVars,
					},
					Env:             createPodEnv(extContext),
					ImagePullPolicy: pullPolicy,
					VolumeMounts:    volumeMounts,
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: extContext.Account,
			Volumes:            volumes,
		},
	}
}

func createPodEnv(executionContext ExecutionContext) []v1.EnvVar {
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
}
