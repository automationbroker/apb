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

	"github.com/automationbroker/apb/pkg/util"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Registry stores a single registry config and references all associated bundle specs
type Registry struct {
	Config registries.Config
	Specs  []*bundle.Spec
}

var defaultLocalOpenShiftConfig = registries.Config{
	Namespaces: []string{"openshift"},
	Name:       "lo-openshift",
	Type:       "local_openshift",
	WhiteList:  []string{".*$"},
}

var defaultDockerHubConfig = registries.Config{
	Name:      "dockerhub",
	Org:       "ansibleplaybookbundle",
	WhiteList: []string{".*-apb$"},
}

var registryConfig Registry
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
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addRegistry(args[0])
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
	registryAddCmd.Flags().StringVar(&registryConfig.Config.Org, "org", "ansibleplaybookbundle", "Type of registry adapter to add")
	registryAddCmd.Flags().StringVar(&registryConfig.Config.URL, "url", "docker.io", "URL of registry adapter to add")
	registryAddCmd.Flags().StringVar(&whitelist, "whitelist", ".*-apb$", "Whitelist for configuration of registry adapter")
	registryConfig.Config.WhiteList = append(registryConfig.Config.WhiteList, whitelist)

	//Registry Remove Flags
	registryRemoveCmd.Flags().StringVar(&removeName, "name", "", "Name of registry adapter to remove")
	registryRemoveCmd.MarkFlagRequired("name")

	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
}

func updateCachedRegistries(regList []Registry) error {
	viper.Set("Registries", regList)
	viper.WriteConfig()
	return nil
}

func addRegistry(regType string) {
	var regList []Registry
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Println("Error unmarshalling config: ", err)
		return
	}

	switch regType:
	case "dockerhub":
		registryConfig.Config = defaultDockerHubConfig{}
	case "local_openshift":
		registryConfig.Config = defaultLocalOpenShiftConfig{}


	for _, reg := range regList {
		if reg.Config.Name == registryConfig.Config.Name {
			fmt.Printf("Error adding registry [%v], found registry with conflicting name [%v]\n", registryConfig.Config.Name, reg.Config.Name)
			return
		}
	}
	regList = append(regList, registryConfig)
	updateCachedRegistries(regList)
	ListImages()
	return
}

func printRegistries(regList []Registry) {
	colName := &util.TableColumn{Header: "NAME"}
	colType := &util.TableColumn{Header: "TYPE"}
	colOrg := &util.TableColumn{Header: "ORG"}
	colURL := &util.TableColumn{Header: "URL"}

	for _, r := range regList {
		colName.Data = append(colName.Data, r.Config.Name)
		colType.Data = append(colType.Data, r.Config.Type)
		colOrg.Data = append(colOrg.Data, r.Config.Org)
		colURL.Data = append(colURL.Data, r.Config.URL)
	}

	tableToPrint := []*util.TableColumn{colName, colType, colOrg, colURL}
	util.PrintTable(tableToPrint)
}

func listRegistries() {
	var regList []Registry
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Printf("Error unmarshalling config: %v", err)
		return
	}
	if len(regList) > 0 {
		fmt.Println("Found registries already in config:")
		printRegistries(regList)
	} else {
		fmt.Println("Found no registries in configuration. Try `apb registry add`.")
	}
	return
}

func removeRegistry() {
	var regList []Registry
	var newRegList []Registry
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Printf("Error unmarshalling config: %v", err)
	}
	for i, r := range regList {
		if r.Config.Name == removeName {
			newRegList = append(regList[:i], regList[i+1:]...)
		}
	}
	updateCachedRegistries(newRegList)
}
