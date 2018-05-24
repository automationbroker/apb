package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var bindingNamespace string

var bindingCmd = &cobra.Command{
	Use:   "binding",
	Short: "Manage bindings",
	Long:  `List, Add, or Delete bindings`,
}

var bindingAddCmd = &cobra.Command{
	Use:   "add <secret name> <app name>",
	Short: "Add a new registry adapter",
	Long:  `Add a new registry adapter to the configuration`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		addBinding(args)
	},
}

var bindingRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a registry adapter",
	Long:  `Remove a registry adapter from stored configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		removeBinding()
	},
}

var bindingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the configured bindings",
	Long:  `List all bindings`,
	Run: func(cmd *cobra.Command, args []string) {
		listBindings()
	},
}

func init() {
	rootCmd.AddCommand(bindingCmd)
	// Binding Add Flags
	bindingAddCmd.Flags().StringVarP(&bindingNamespace, "namespace", "p", "", "Namespace of binding")
	bindingAddCmd.MarkFlagRequired("namespace")

	bindingCmd.AddCommand(bindingAddCmd)
	bindingCmd.AddCommand(bindingListCmd)
	bindingCmd.AddCommand(bindingRemoveCmd)
}

/*
func updateCachedRegistries(registries []registries.Config) error {
	viper.Set("Registries", registries)
	viper.WriteConfig()
	return nil
}
*/

func addBinding(args []string) {
	fmt.Println("addBindings called")
	secretName := args[0]
	newSecretName := fmt.Sprintf("%v-creds", secretName)
	appName := args[1]
	fmt.Println(secretName)
	fmt.Println(appName)
	fmt.Printf("Create a binding using secret [%s] to app [%s]\n", secretName, appName)
	secretData, err := extractCredentialsAsSecret(secretName, bindingNamespace)
	if err != nil {
		fmt.Errorf("Unable to retrieve secret data from secret [%v]\n", err)
		return
	}
	extCreds, err := buildExtractedCredentials(secretData)
	if err != nil {
		fmt.Errorf("Unexpected error building extracted creds: %v\n", err)
	}
	data := map[string][]byte{}
	for key, value := range extCreds.Credentials {
		data[key] = []byte(value.(string))
	}
	s := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: newSecretName,
		},
		Data: data,
	}

	k8scli, err := clients.Kubernetes()
	if err != nil {
		panic(err.Error())
	}
	_, err = k8scli.Client.CoreV1().Secrets(bindingNamespace).Create(s)
	if err != nil {
		fmt.Errorf("Unable to create secret %v in namespace %v", newSecretName, bindingNamespace)
		return
	}

	fmt.Printf("Successfully created secret [%v] in namespace [%v].\n", newSecretName, bindingNamespace)
	fmt.Printf("Type the following command to attach the binding to your application:\n")
	fmt.Printf("oc set env dc/%v --from=secret/%v\n", appName, newSecretName)
	return

}

func listBindings() {
	fmt.Println("listBindings called")
}

func removeBinding() {
	fmt.Println("removeBinding called")
}

// ExtractCredentialsAsSecret - Extract credentials from APB as secret in namespace.
func extractCredentialsAsSecret(podname string, namespace string) ([]byte, error) {
	k8s, err := clients.Kubernetes()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrive kubernetes client - %v", err)
	}

	secret, err := k8s.GetSecretData(podname, namespace)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve secret [ %v ] - %v", podname, err)
	}

	return secret["fields"], nil
}

func buildExtractedCredentials(output []byte) (*bundle.ExtractedCredentials, error) {
	creds := make(map[string]interface{})
	err := json.Unmarshal(output, &creds)
	if err != nil {
		return nil, err
	}
	return &bundle.ExtractedCredentials{Credentials: creds}, nil
}
