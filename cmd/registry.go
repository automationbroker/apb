package cmd

import (
	"fmt"

	"github.com/automationbroker/bundle-lib/registries"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var registryConfig registries.Config

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Configure registry adapters",
	Long:  `List, Add, or Delete registry adapters from configuration`,
}

var registryAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new registry adapter",
	Long:  `Add a new registry adapter to the configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		addRegistry()
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)
	registryAddCmd.Flags().StringVar(&registryConfig.Type, "type", "dockerhub", "Type of registry adapter to add")
	registryAddCmd.Flags().StringVar(&registryConfig.Org, "org", "ansibleplaybookbundle", "Type of registry adapter to add")
	registryAddCmd.Flags().StringVar(&registryConfig.URL, "url", "docker.io", "URL of registry adapter to add")
	registryAddCmd.Flags().StringVar(&registryConfig.Name, "name", "docker", "Name of registry adapter to add")
	registryCmd.AddCommand(registryAddCmd)
}

func updateCachedRegistries(registries []registries.Config) error {
	viper.Set("Registries", registries)
	viper.WriteConfig()
	return nil
}

func addRegistry() {
	reg, err := registries.NewRegistry(registryConfig, "ansible-service-broker")
	if err != nil {
		fmt.Printf("Error creating new registry adapter: %v", err)
		return
	}
	currentReg := viper.Get("Registries")
	fmt.Printf("Current: %v", currentReg)
	fmt.Printf("new: %v", reg)
	return
}

func listRegistries() {
	var registries []*registries.Config = nil
	err := viper.UnmarshalKey("Registries", &registries)
	if err != nil {
		fmt.Println("Error unmarshalling config: ", err)
		return
	}
	if len(registries) > 0 {
		fmt.Println("Found registries already in config")
		for _, r := range registries {
			fmt.Printf("%v - %v\n", r.Name, r.URL)
		}
		return
	}
}
