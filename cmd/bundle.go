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
	"github.com/automationbroker/sbcli/pkg/runner"
	"github.com/automationbroker/sbcli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// Refresh - indicates whether we should refresh the list of images
var Refresh bool

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Interact with ServiceBundles",
	Long:  `List and execute ServiceBundles from configured registry adapters`,
}

var bundleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ServiceBundle images",
	Long:  `List ServiceBundles from a registry adapter`,
	Run: func(cmd *cobra.Command, args []string) {
		listImages()
	},
}

var bundleNamespace string

var bundleProvisionCmd = &cobra.Command{
	Use:   "provision <bundle name>",
	Short: "Provision ServiceBundle images",
	Long:  `Provision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunBundle("provision", bundleNamespace, args)
	},
}

var bundleDeprovisionCmd = &cobra.Command{
	Use:   "deprovision <bundle name>",
	Short: "Deprovision ServiceBundle images",
	Long:  `Deprovision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunBundle("deprovision", bundleNamespace, args)
	},
}

func init() {
	rootCmd.AddCommand(bundleCmd)

	bundleListCmd.Flags().BoolVarP(&Refresh, "refresh", "r", false, "refresh list of specs")
	bundleCmd.AddCommand(bundleListCmd)

	bundleProvisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "p", "", "Namespace to provision bundle to")
	bundleProvisionCmd.MarkFlagRequired("namespace")
	bundleCmd.AddCommand(bundleProvisionCmd)

	bundleDeprovisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "p", "", "Namespace to provision bundle to")
	bundleDeprovisionCmd.MarkFlagRequired("namespace")
	bundleCmd.AddCommand(bundleDeprovisionCmd)
}

// Get images from a single registry
func getImages(registryMetadata Registry) ([]*bundle.Spec, error) {
	var specList []*bundle.Spec

	authNamespace := ""
	registry, err := registries.NewRegistry(registryMetadata.Config, authNamespace)
	if err != nil {
		log.Error("Error from creating a NewRegistry")
		log.Error(err)
		return nil, err
	}

	specs, count, err := registry.LoadSpecs()
	if err != nil {
		log.Errorf("registry: %v was unable to complete bootstrap - %v",
			registry.RegistryName(), err)
		return nil, err
	}
	log.Infof("Registry %v has %d bundles available from %d images scanned", registry.RegistryName(), len(specs), count)

	specList = append(specList, specs...)
	return specList, nil
}

func printRegConfigSpecs(regConfigs []Registry) {
	colFQName := &util.TableColumn{Header: "BUNDLE"}
	colImage := &util.TableColumn{Header: "IMAGE"}
	colRegName := &util.TableColumn{Header: "REGISTRY"}

	for _, r := range regConfigs {
		for _, s := range r.Specs {
			colFQName.Data = append(colFQName.Data, s.FQName)
			colImage.Data = append(colImage.Data, s.Image)
			colRegName.Data = append(colRegName.Data, r.Config.Name)
		}
	}

	tableToPrint := []*util.TableColumn{colFQName, colImage, colRegName}
	util.PrintTable(tableToPrint)
}

func listImages() {
	var regConfigs []Registry
	var newRegConfigs []Registry

	err := viper.UnmarshalKey("Registries", &regConfigs)
	if err != nil {
		log.Error("Error unmarshalling config: ", err)
		return
	}

	for _, regConfig := range regConfigs {
		if len(regConfig.Specs) > 0 && Refresh == false {
			fmt.Printf("Found specs already in registry: [%s]\n", regConfig.Config.Name)
			newRegConfigs = append(newRegConfigs, regConfig)
			continue
		}
		fmt.Printf("Getting specs for registry: [%s]\n", regConfig.Config.Name)
		specs, err := getImages(regConfig)
		if err != nil {
			log.Error("Error getting images")
			continue
		}

		regConfig.Specs = specs
		newRegConfigs = append(newRegConfigs, regConfig)
	}
	printRegConfigSpecs(newRegConfigs)

	err = updateCachedRegistries(newRegConfigs)
	if err != nil {
		log.Error("Error updating cache")
		return
	}
}
