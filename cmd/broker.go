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

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/automationbroker/sbcli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var brokerName string

var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "Interact with an Automation Broker instance",
	Long:  `List or Bootstrap bundles on an Automation Broker instance`,
}

var brokerCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "List available service bundles in broker catalog",
	Long:  `Fetch list of service bundles in Automation Broker catalog`,
	Run: func(cmd *cobra.Command, args []string) {
		listBrokerCatalog()
	},
}

var brokerBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap an Automation Broker instance",
	Long:  `Fetch list of service bundles in Automation Broker catalog`,
	Run: func(cmd *cobra.Command, args []string) {
		listBroker()
	},
}

func init() {
	rootCmd.AddCommand(brokerCmd)
	// Broker List Flags
	brokerCmd.Flags().StringVar(&brokerName, "name", "automation-broker", "Name of Automation Broker instance")

	//Broker Bootstrap Flags
	brokerBootstrap.Flags().StringVar(&removeName, "name", "", "Name of registry adapter to remove")
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

func addRegistry() {
	var regList []Registry
	err := viper.UnmarshalKey("Registries", &regList)
	if err != nil {
		fmt.Println("Error unmarshalling config: ", err)
		return
	}
	for _, reg := range regList {
		if reg.Config.Name == registryConfig.Config.Name {
			fmt.Printf("Error adding registry [%v], found registry with conflicting name [%v]\n", registryConfig.Config.Name, reg.Config.Name)
			return
		}
	}
	regList = append(regList, registryConfig)
	updateCachedRegistries(regList)
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
		fmt.Println("Found no registries in configuration. Try `sbcli registry add`.")
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
