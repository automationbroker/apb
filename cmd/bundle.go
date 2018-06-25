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
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
	"github.com/automationbroker/sbcli/pkg/runner"
	"github.com/automationbroker/sbcli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// Refresh - indicates whether we should refresh the list of images
var Refresh bool

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Interact with ServiceBundles",
	Long:  `List, execute and build ServiceBundles`,
}

var bundleMetadataFilename string
var containerMetadataFilename string

var bundlePrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Stamp ServiceBundle metadata onto Dockerfile as b64",
	Long:  `Prepare for ServiceBundle image build by stamping apb.yml contents onto Dockerfile as b64`,
	Run: func(cmd *cobra.Command, args []string) {
		stampBundleMetadata(bundleMetadataFilename, containerMetadataFilename)
	},
}

var bundleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ServiceBundle images",
	Long:  `List ServiceBundles from a registry adapter`,
	Run: func(cmd *cobra.Command, args []string) {
		ListImages()
	},
}

var bundleRegistry string

var bundleInfoCmd = &cobra.Command{
	Use:   "info <bundle name>",
	Short: "Print info on ServiceBundle image",
	Long:  `Print metadata, plans, and params associated with ServiceBundle image`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		showBundleInfo(args[0], bundleRegistry)
	},
}

var bundleNamespace string
var sandboxRole string

var bundleProvisionCmd = &cobra.Command{
	Use:   "provision <bundle name>",
	Short: "Provision ServiceBundle images",
	Long:  `Provision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunBundle("provision", bundleNamespace, args[0], sandboxRole, bundleRegistry, args[1:])
	},
}

var bundleDeprovisionCmd = &cobra.Command{
	Use:   "deprovision <bundle name>",
	Short: "Deprovision ServiceBundle images",
	Long:  `Deprovision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunBundle("deprovision", bundleNamespace, args[0], sandboxRole, bundleRegistry, args[1:])
	},
}

func init() {
	rootCmd.AddCommand(bundleCmd)

	bundlePrepareCmd.Flags().StringVarP(&bundleMetadataFilename, "bmeta", "b", "apb.yml", "Bundle metadata file to encode as b64")
	bundlePrepareCmd.Flags().StringVarP(&containerMetadataFilename, "cmeta", "c", "Dockerfile", "Container metadata file to stamp")
	bundleCmd.AddCommand(bundlePrepareCmd)

	bundleListCmd.Flags().BoolVarP(&Refresh, "refresh", "r", false, "refresh list of specs")
	bundleCmd.AddCommand(bundleListCmd)

	bundleInfoCmd.Flags().StringVarP(&bundleRegistry, "registry", "r", "", "Registry to retrieve bundle info from")
	bundleCmd.AddCommand(bundleInfoCmd)

	bundleProvisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "n", "", "Namespace to provision bundle to")
	bundleProvisionCmd.Flags().StringVarP(&sandboxRole, "role", "r", "edit", "ClusterRole to be applied to Bundle sandbox")
	bundleProvisionCmd.Flags().StringVarP(&bundleRegistry, "registry", "", "", "Registry to load bundle from")
	bundleProvisionCmd.MarkFlagRequired("namespace")
	bundleCmd.AddCommand(bundleProvisionCmd)

	bundleDeprovisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "n", "", "Namespace to provision bundle to")
	bundleDeprovisionCmd.Flags().StringVarP(&sandboxRole, "role", "r", "edit", "ClusterRole to be applied to Bundle sandbox")
	bundleDeprovisionCmd.Flags().StringVarP(&bundleRegistry, "registry", "", "", "Registry to load bundle from")
	bundleDeprovisionCmd.MarkFlagRequired("namespace")
	bundleCmd.AddCommand(bundleDeprovisionCmd)
}

// ListImages finds and prints inforomation on bundle images from all the registries
func ListImages() {
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
			log.Errorf("Error getting images - %v", err)
			continue
		}

		regConfig.Specs = specs
		newRegConfigs = append(newRegConfigs, regConfig)
	}
	printRegConfigSpecs(newRegConfigs)

	err = updateCachedRegistries(newRegConfigs)
	if err != nil {
		log.Errorf("Error updating cache - %v", err)
		return
	}
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

func printBundleInfo(bundleSpec *bundle.Spec) {
	fmt.Printf(" %-11s  |  %v\n", "NAME", bundleSpec.FQName)
	fmt.Printf(" %-11s  |  %v\n", "DESCRIPTION", bundleSpec.Description)
	fmt.Printf(" %-11s  |  %v\n", "IMAGE", bundleSpec.Image)
	fmt.Printf(" %-11s  |  %v\n", "ASYNC BIND", bundleSpec.Async)
	fmt.Printf(" %-11s  |  %v\n", "BINDABLE", bundleSpec.Bindable)
	fmt.Printf(" %-11s  |  %v\n", "VERSION", bundleSpec.Version)
	fmt.Printf(" %-11s  |  %v\n", "APB RUNTIME", bundleSpec.Runtime)
	fmt.Printf(" %-11s  | \n", "")

	for i, plan := range bundleSpec.Plans {
		fmt.Printf(" %-11s  |  %v\n", "PLAN", plan.Name)
		for _, param := range plan.Parameters {
			fmt.Printf("   %-9s  |    %v\n", "param", param.Name)
		}
		if i < len(bundleSpec.Plans)-1 {
			fmt.Printf(" %-11s  | \n", "")
		}
	}
	fmt.Println()
}

func showBundleInfo(bundleName string, registryName string) {
	var regConfigs []Registry

	err := viper.UnmarshalKey("Registries", &regConfigs)
	if err != nil {
		log.Error("Error unmarshalling config: ", err)
		return
	}

	var bundleSpecMatches []*bundle.Spec

	for _, regConfig := range regConfigs {
		if len(registryName) > 0 && regConfig.Config.Name != registryName {
			continue
		}
		for _, bundleSpec := range regConfig.Specs {
			if bundleSpec.FQName == bundleName {
				bundleSpecMatches = append(bundleSpecMatches, bundleSpec)
				fmt.Printf("Found bundle [%v] in registry: [%v]\n", bundleName, regConfig.Config.Name)
			}
		}
	}

	if len(bundleSpecMatches) == 0 {
		log.Errorf("No bundles found with name [%v]", bundleName)
		return
	}
	if len(bundleSpecMatches) > 1 {
		log.Warnf("Found multiple bundles matching name [%v]. Specify a registry with -r or --registry.", bundleName)
		return
	}
	fmt.Println()
	printBundleInfo(bundleSpecMatches[0])

	return
}

// stampBundleMetadata encodes bundle metadata (e.g. apb.yml) as b64 and stamps onto container build file (e.g. Dockerfile)
func stampBundleMetadata(bundleMetaFilename string, containerMetaFilename string) {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Errorf("Couldn't determine working directory")
		return
	}

	bundleMetaPath := filepath.Join(workingDir, bundleMetaFilename)
	containerMetaPath := filepath.Join(workingDir, containerMetaFilename)

	bundleMeta, errB := ioutil.ReadFile(bundleMetaPath)
	if errB != nil {
		log.Errorf("Bundle metadata file [%s] not found in working directory", bundleMetaFilename)
	}
	containerMeta, errC := ioutil.ReadFile(containerMetaPath)
	if errC != nil {
		log.Errorf("Container metadata file [%s] not found in working directory", containerMetaFilename)
	}
	if errB != nil || errC != nil {
		return
	}

	newContainerMeta := addBundleMetadata(bundleMeta, containerMeta)
	if len(newContainerMeta) == 0 {
		log.Errorf("Failed to add [%s] contents to [%s]", bundleMetaFilename, containerMetaFilename)
		return
	}

	writeErr := ioutil.WriteFile(containerMetaFilename, newContainerMeta, 0644)
	if writeErr != nil {
		log.Errorf("Container metadata file [%s] could not be written", containerMetaFilename)
		return
	}
	fmt.Printf("Wrote b64 encoded [%s] to [%s]\n", bundleMetaFilename, containerMetaFilename)
}

func addBundleMetadata(bundleMeta []byte, containerMeta []byte) []byte {
	lineBreakAfter := 76
	lineBreakText := []byte("\\\n")

	// Encode bundle metadata as b64 and add linebreaks
	bundleMetaEncoded := make([]byte, base64.StdEncoding.EncodedLen(len(bundleMeta)))
	base64.StdEncoding.Encode(bundleMetaEncoded, bundleMeta)
	bundleMetaEncoded = addLineBreaks(bundleMetaEncoded, lineBreakText, lineBreakAfter)

	// Match the apb spec "LABEL" line up to the opening quote on the b64 blob
	labelRegexp, _ := regexp.Compile(`LABEL.*(\")?com\.redhat\.apb\.spec(\")?\=(\\)?\n?\"`)
	indices := labelRegexp.FindIndex(containerMeta)

	if len(indices) == 0 {
		log.Errorf("Didn't find expected bundle label in container metadata file")
		return []byte{}
	}
	lineStartIndex := indices[0]
	blobStartIndex := indices[1]
	blobEndOffset := bytes.IndexByte(containerMeta[blobStartIndex:], byte('"'))
	if blobEndOffset == -1 {
		log.Errorf("Didn't find end of bundle label in container metadata file")
		return []byte{}
	}
	blobEndIndex := blobStartIndex + blobEndOffset

	// Build new label section
	bundleMetaSection := containerMeta[lineStartIndex : blobStartIndex-1]
	if !bytes.Contains(bundleMetaSection, []byte(lineBreakText)) {
		bundleMetaSection = append(bundleMetaSection, lineBreakText...)
	}
	bundleMetaSection = append(bundleMetaSection, byte('"'))
	bundleMetaSection = append(bundleMetaSection, bundleMetaEncoded...)

	// Build final output with linebreaked "LABEL" b64 content
	newContainerMeta := containerMeta[0:lineStartIndex]
	newContainerMeta = append(newContainerMeta, []byte(bundleMetaSection)...)
	newContainerMeta = append(newContainerMeta, containerMeta[blobEndIndex:]...)

	return newContainerMeta
}

func addLineBreaks(text []byte, lineBreakText []byte, breakAfter int) []byte {
	newText := []byte("")
	for index, currentByte := range text {
		if currentByte != '\n' {
			newText = append(newText, currentByte)
		}
		if (index+1)%breakAfter == 0 {
			newText = append(newText, lineBreakText...)
		}
	}
	return newText
}
