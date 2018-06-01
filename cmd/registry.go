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
	"fmt"

	"github.com/automationbroker/bundle-lib/registries"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var registryConfig registries.Config
var whitelist string
var removeName string

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

var registryRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a registry adapter",
	Long:  `Remove a registry adapter from stored configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		removeRegistry()
	},
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the configured registry adapters",
	Long:  `List all registry adapters in the configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		listRegistries()
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)
	// Registry Add Flags
	registryAddCmd.Flags().StringVar(&registryConfig.Type, "type", "dockerhub", "Type of registry adapter to add")
	registryAddCmd.Flags().StringVar(&registryConfig.Org, "org", "ansibleplaybookbundle", "Type of registry adapter to add")
	registryAddCmd.Flags().StringVar(&registryConfig.URL, "url", "docker.io", "URL of registry adapter to add")
	registryAddCmd.Flags().StringVar(&registryConfig.Name, "name", "docker", "Name of registry adapter to add")
	registryAddCmd.Flags().StringVar(&whitelist, "whitelist", ".*-apb$", "Whitelist for configuration of registry adapter")
	registryConfig.WhiteList = append(registryConfig.WhiteList, whitelist)

	//Registry Remove Flags
	registryRemoveCmd.Flags().StringVar(&removeName, "name", "", "Name of registry adapter to remove")
	registryRemoveCmd.MarkFlagRequired("name")

	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
}

func updateCachedRegistries(registries []registries.Config) error {
	viper.Set("Registries", registries)
	viper.WriteConfig()
	return nil
}

func addRegistry() {
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

func removeRegistry() {
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
