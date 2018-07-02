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

// Default registry configuration section
var defaultLocalOpenShiftConfig = registries.Config{
	Namespaces: []string{"openshift"},
	Name:       "lo-openshift",
	Type:       "local_openshift",
	WhiteList:  []string{".*$"},
}

var defaultDockerHubConfig = registries.Config{
	Name:      "dockerhub",
	Type:      "dockerhub",
	URL:       "docker.io",
	Org:       "ansibleplaybookbundle",
	WhiteList: []string{".*-apb$"},
}

var defaultHelmConfig = registries.Config{
	Name:      "helm",
	Type:      "helm",
	URL:       "https://kubernetes-charts.storage.googleapis.com",
	Runner:    "docker.io/automationbroker/helm-runner:latest",
	WhiteList: []string{".*$"},
}

// Registry cmd vars
var registryConfig Registry

// Registry add flags
var nsList []string
var whitelist []string
var regOrg string
var regUrl string
var addName string

// Registry commands
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Configure registry adapters",
	Long:  `List, Add, or Delete registry adapters from configuration`,
}

var registryAddCmd = &cobra.Command{
	Use:   "add <registry_type>",
	Short: "Add a new registry adapter",
	Long:  `Add a new registry adapter to the configuration`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addRegistry(args[0])
	},
}

var registryRemoveCmd = &cobra.Command{
	Use:   "remove <registry_name>",
	Short: "Remove a registry adapter",
	Args:  cobra.MinimumNArgs(1),
	Long:  `Remove a registry adapter from stored configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		removeRegistry(args[0])
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
	registryAddCmd.Flags().StringVar(&regOrg, "org", "", "Type of registry adapter to add")
	registryAddCmd.Flags().StringVar(&regUrl, "url", "", "URL of registry adapter to add")
	registryAddCmd.Flags().StringSliceVar(&whitelist, "whitelist", []string{}, "Comma-separated whitelist for configuration of registry adapter")
	registryAddCmd.Flags().StringSliceVar(&nsList, "namespace", []string{}, "Comma-separated list of namespaces to configure local_openshift adapter")
	registryAddCmd.Flags().StringVar(&addName, "name", "", "Name of registry adapter to add")

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

	// TODO: Add all cases here. For simplicity I'm keeping it like this until the v2 work merges
	switch regType {
	case "dockerhub":
		registryConfig.Config = defaultDockerHubConfig
	case "local_openshift":
		registryConfig.Config = defaultLocalOpenShiftConfig
	case "helm":
		registryConfig.Config = defaultHelmConfig
	default:
		fmt.Printf("Unrecognized type of registry [%v]\n", regType)
		return
	}

	registryConfig.Config = applyOverrides(registryConfig.Config)

	for _, reg := range regList {
		if reg.Config.Name == registryConfig.Config.Name {
			fmt.Printf("Error adding registry [%v], found registry with conflicting name [%v]. Try specifying a name with --name flag.\n", registryConfig.Config.Name, reg.Config.Name)
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

func applyOverrides(conf registries.Config) registries.Config {
	if regOrg != "" {
		conf.Org = regOrg
	}
	if regUrl != "" {
		conf.URL = regUrl
	}
	if len(nsList) > 0 {
		conf.Namespaces = nsList
	}
	if len(whitelist) > 0 {
		conf.WhiteList = whitelist
	}
	if addName != "" {
		conf.Name = addName
	}
	return conf
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

func removeRegistry(name string) {
	var regList []Registry
	var newRegList []Registry
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Printf("Error unmarshalling config: %v", err)
	}
	for i, r := range regList {
		if r.Config.Name == name {
			fmt.Printf("Found registry [%v]. Removing from list.\n", name)
			newRegList = append(regList[:i], regList[i+1:]...)
			updateCachedRegistries(newRegList)
			return
		}
	}
	fmt.Printf("Failed to remove registry [%v]. Check the spelling and try again.\n", name)
}
