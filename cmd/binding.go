package cmd

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/registries"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var whitelist string
var removeName string

var bindingCmd = &cobra.Command{
	Use:   "binding",
	Short: "Manage bindings",
	Long:  `List, Add, or Delete bindings`,
}

var bindingAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new registry adapter",
	Long:  `Add a new registry adapter to the configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		addBinding()
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
		listRegistries()
	},
}

func init() {
	rootCmd.AddCommand(bindingCmd)
	// Binding Add Flags
	/*bindingAddCmd.Flags().StringVar(&registryConfig.Type, "type", "dockerhub", "Type of registry adapter to add")
	bindingAddCmd.Flags().StringVar(&registryConfig.Org, "org", "ansibleplaybookbundle", "Type of registry adapter to add")
	bindingAddCmd.Flags().StringVar(&registryConfig.URL, "url", "docker.io", "URL of registry adapter to add")
	bindingAddCmd.Flags().StringVar(&registryConfig.Name, "name", "docker", "Name of registry adapter to add")
	bindingAddCmd.Flags().StringVar(&whitelist, "whitelist", ".*-apb$", "Whitelist for configuration of registry adapter")
	*/

	// Binding Remove Flags
	bindingRemoveCmd.Flags().StringVar(&removeName, "name", "", "Name of registry adapter to remove")
	bindingRemoveCmd.MarkFlagRequired("name")

	bindingCmd.AddCommand(bindingAddCmd)
	bindingCmd.AddCommand(bindingListCmd)
	bindingCmd.AddCommand(bindingRemoveCmd)
}

func updateCachedRegistries(registries []registries.Config) error {
	viper.Set("Registries", registries)
	viper.WriteConfig()
	return nil
}

func addBinding() {
	var regList []registries.Config
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Println("Error unmarshalling config: ", err)
		return
	}
	for _, reg := range regList {
		if reg.Name == registryConfig.Name {
			fmt.Printf("Error adding registry [%v], found registry with conflicting name [%v]\n", registryConfig.Name, reg.Name)
			return
		}
	}

	regList = append(regList, registryConfig)
	updateCachedRegistries(regList)
	return
}

func listRegistries() {
	var regList []registries.Config
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Printf("Error unmarshalling config: %v", err)
		return
	}
	if len(regList) > 0 {
		fmt.Println("Found registries already in config:")
		for _, r := range regList {
			fmt.Printf("name: %v - type: %v - organization: %v - URL: %v\n", r.Name, r.Type, r.Org, r.URL)
		}
	} else {
		fmt.Println("Found no registries in configuration. Try `sbcli registry add`.")
	}
	return
}

func removeBinding() {
	var regList []registries.Config
	var newRegList []registries.Config
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Printf("Error unmarshalling config: %v", err)
	}
	for i, r := range regList {
		if r.Name == removeName {
			newRegList = append(regList[:i], regList[i+1:]...)
		}
	}
	updateCachedRegistries(newRegList)
}
