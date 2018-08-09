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

	"github.com/automationbroker/apb/pkg/config"
	"github.com/automationbroker/apb/pkg/util"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/spf13/cobra"
)

// Default registry configuration section
var defaultLocalOpenShiftConfig = registries.Config{
	Namespaces: []string{"openshift"},
	Name:       "local",
	Type:       "local_openshift",
	WhiteList:  []string{".*$"},
}

var defaultQuayConfig = registries.Config{
	Name:      "quay",
	Type:      "quay",
	Org:       "redhat",
	URL:       "http://quay.io",
	WhiteList: []string{".*$"},
}

var defaultDockerHubConfig = registries.Config{
	Name:      "dockerhub",
	Type:      "dockerhub",
	URL:       "docker.io",
	Org:       "ansibleplaybookbundle",
	WhiteList: []string{".*$"},
}

var defaultHelmConfig = registries.Config{
	Name:      "helm",
	Type:      "helm",
	URL:       "https://kubernetes-charts.storage.googleapis.com",
	Runner:    "docker.io/automationbroker/helm-runner:latest",
	WhiteList: []string{".*$"},
}

// Registry cmd vars
var registryConfig config.Registry

// Registry commands
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Configure registry adapters",
	Long:  `List, Add, or Delete registry adapters from configuration`,
}

var registryAddCmd = &cobra.Command{
	Use:   "add <registry_name>",
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
	registryAddCmd.Flags().StringVarP(&registryConfig.Config.Type, "type", "t", "dockerhub", "registry type (dockerhub, local_openshift, helm)")
	registryAddCmd.Flags().StringVar(&registryConfig.Config.Tag, "tag", "", "specify the tag of images in registry (e.g. 'latest')")
	registryAddCmd.Flags().StringVar(&registryConfig.Config.Org, "org", "", "organization for 'dockerhub' adapter to search (e.g. 'ansible-playbook-bundle')")
	registryAddCmd.Flags().StringVar(&registryConfig.Config.URL, "url", "", "URL (e.g. docker.io)")
	registryAddCmd.Flags().StringSliceVar(&registryConfig.Config.WhiteList, "whitelist", []string{}, "regexes for filtering registry contents (e.g. '.*apb$,.*bundle$')")
	registryAddCmd.Flags().StringSliceVar(&registryConfig.Config.Namespaces, "namespaces", []string{}, "namespaces for 'local_openshift' adapter to search  (e.g. 'openshift,my-project')")

	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryRemoveCmd)
}

func addRegistry(addName string) {
	var regList []config.Registry
	var newConfig config.Registry
	err := config.Registries.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Println("Error unmarshalling config: ", err)
		return
	}

	// TODO: Add all cases here. For simplicity I'm keeping it like this until the v2 work merges
	switch registryConfig.Config.Type {
	case "dockerhub":
		newConfig.Config = defaultDockerHubConfig
	case "local_openshift":
		newConfig.Config = defaultLocalOpenShiftConfig
	case "helm":
		newConfig.Config = defaultHelmConfig
	case "quay":
		newConfig.Config = defaultQuayConfig
	default:
		fmt.Printf("Unrecognized registry type [%v]\n", registryConfig.Config.Type)
		fmt.Printf("Supported types are: dockerhub, local_openshift, helm.\n")
		return
	}
	newConfig.Config.Name = addName

	applyOverrides(&newConfig.Config, registryConfig.Config)

	for _, reg := range regList {
		if reg.Config.Name == newConfig.Config.Name {
			fmt.Printf("Error adding registry [%v], found registry with conflicting name [%v]. Try specifying a different name.\n", registryConfig.Config.Name, reg.Config.Name)
			return
		}
	}
	regList = append(regList, newConfig)
	config.UpdateCachedRegistries(config.Registries, regList)
	ListImages()
	return
}

func printRegistries(regList []config.Registry) {
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

func applyOverrides(conf *registries.Config, params registries.Config) {
	if params.Org != "" {
		conf.Org = params.Org
	}
	if params.URL != "" {
		conf.URL = params.URL
	}
	if params.Tag != "" {
		conf.Tag = params.Tag
	}
	if len(params.Namespaces) > 0 {
		conf.Namespaces = params.Namespaces
	}
	if len(params.WhiteList) > 0 {
		conf.WhiteList = params.WhiteList
	}
}

func listRegistries() {
	var regList []config.Registry
	err := config.Registries.UnmarshalKey("Registries", &regList)
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
	var regList []config.Registry
	var newRegList []config.Registry
	err := config.Registries.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Printf("Error unmarshalling config: %v", err)
	}
	for i, r := range regList {
		if r.Config.Name == name {
			fmt.Printf("Found registry [%v]. Removing from list.\n", name)
			newRegList = append(regList[:i], regList[i+1:]...)
			config.UpdateCachedRegistries(config.Registries, newRegList)
			return
		}
	}
	fmt.Printf("Failed to remove registry [%v]. Check the spelling and try again.\n", name)
}
