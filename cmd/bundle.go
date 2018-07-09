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

	"github.com/automationbroker/apb/pkg/runner"
	"github.com/automationbroker/apb/pkg/util"
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
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
var noLineBreaks bool

var bundlePrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Stamp ServiceBundle metadata onto Dockerfile as b64",
	Long:  `Prepare for ServiceBundle image build by stamping apb.yml contents onto Dockerfile as b64`,
	Run: func(cmd *cobra.Command, args []string) {
		stampBundleMetadata(bundleMetadataFilename, containerMetadataFilename, noLineBreaks)
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
var kubeConfig string
var printLogs bool

var bundleProvisionCmd = &cobra.Command{
	Use:   "provision <bundle name>",
	Short: "Provision ServiceBundle images",
	Long:  `Provision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		executeBundle("provision", args)
	},
}

var bundleDeprovisionCmd = &cobra.Command{
	Use:   "deprovision <bundle name>",
	Short: "Deprovision ServiceBundle images",
	Long:  `Deprovision ServiceBundles from a registry adapter`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		executeBundle("deprovision", args)
	},
}

func init() {
	bundleCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to kubeconfig to use")
	rootCmd.AddCommand(bundleCmd)

	bundlePrepareCmd.Flags().StringVarP(&bundleMetadataFilename, "bundlemeta", "b", "apb.yml", "Bundle metadata file to encode as b64")
	bundlePrepareCmd.Flags().StringVarP(&containerMetadataFilename, "containermeta", "c", "Dockerfile", "Container metadata file to stamp")
	bundlePrepareCmd.Flags().BoolVarP(&noLineBreaks, "nolinebreak", "n", false, "Skip adding linebreaks to b64 bundle spec")
	bundleCmd.AddCommand(bundlePrepareCmd)

	bundleListCmd.Flags().BoolVarP(&Refresh, "refresh", "r", false, "refresh list of specs")
	bundleCmd.AddCommand(bundleListCmd)

	bundleInfoCmd.Flags().StringVarP(&bundleRegistry, "registry", "r", "", "Registry to retrieve bundle info from")
	bundleCmd.AddCommand(bundleInfoCmd)

	bundleProvisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "n", "", "Namespace to provision bundle to")
	bundleProvisionCmd.Flags().StringVarP(&sandboxRole, "role", "r", "edit", "ClusterRole to be applied to Bundle sandbox")
	bundleProvisionCmd.Flags().StringVarP(&bundleRegistry, "registry", "", "", "Registry to load bundle from")
	bundleProvisionCmd.Flags().BoolVarP(&printLogs, "follow", "f", false, "Print logs from provision pod")
	bundleCmd.AddCommand(bundleProvisionCmd)

	bundleDeprovisionCmd.Flags().StringVarP(&bundleNamespace, "namespace", "n", "", "Namespace to deprovision bundle from")
	bundleDeprovisionCmd.Flags().StringVarP(&sandboxRole, "role", "r", "edit", "ClusterRole to be applied to Bundle sandbox")
	bundleDeprovisionCmd.Flags().StringVarP(&bundleRegistry, "registry", "", "", "Registry to load bundle from")
	bundleDeprovisionCmd.Flags().BoolVarP(&printLogs, "follow", "f", false, "Print logs from deprovision pod")
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

func executeBundle(action string, args []string) {
	if bundleNamespace == "" {
		bundleNamespace = util.GetCurrentNamespace(kubeConfig)
		if bundleNamespace == "" {
			log.Errorf("Failed to get current namespace. Try supplying it with --namespace.")
			return
		}
	}
	log.Debugf("Running bundle [%v] with action [%v] in namespace [%v].", args[0], action, bundleNamespace)
	runner.RunBundle(action, bundleNamespace, args[0], sandboxRole, bundleRegistry, printLogs, args[1:])
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

func stampBundleMetadata(bundleMetaFilename string, containerMetaFilename string, noLineBreaks bool) {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Errorf("Couldn't determine working directory")
		return
	}

	bundleMetaPath := filepath.Join(workingDir, bundleMetaFilename)
	containerMetaPath := filepath.Join(workingDir, containerMetaFilename)

	fileErr := false
	bundleMeta, err := ioutil.ReadFile(bundleMetaPath)
	if err != nil {
		log.Errorf("Bundle metadata file [%s] not found in working directory", bundleMetaFilename)
		fileErr = true
	}
	containerMeta, err := ioutil.ReadFile(containerMetaPath)
	if err != nil {
		log.Errorf("Container metadata file [%s] not found in working directory", containerMetaFilename)
		fileErr = true
	}
	if fileErr {
		return
	}

	newContainerMeta := addBundleMetadata(bundleMeta, containerMeta, noLineBreaks)
	if len(newContainerMeta) == 0 {
		log.Errorf("Failed to add [%s] contents to [%s]", bundleMetaFilename, containerMetaFilename)
		return
	}

	writeErr := ioutil.WriteFile(containerMetaFilename, newContainerMeta, 0644)
	if writeErr != nil {
		log.Errorf("Container metadata file [%s] could not be written", containerMetaFilename)
		return
	}
	fmt.Printf("Wrote b64 encoded [%s] to [%s]\n\n", bundleMetaFilename, containerMetaFilename)

	buildConfigCmd := `oc new-build . --to <bundle-name>`
	fmt.Printf("Create a buildconfig:\n%s\n\n", buildConfigCmd)

	buildTriggerCmd := `oc start-build --from-dir . <bundle-name>`
	fmt.Printf("Start a build:\n%s\n\n", buildTriggerCmd)
}

func addBundleMetadata(bMeta []byte, cMeta []byte, noLineBreaks bool) []byte {
	lineBreakAfter := 76
	lineBreakText := []byte("\\\n")

	// Encode bundle metadata as b64 and add linebreaks
	bMetaEncoded := make([]byte, base64.StdEncoding.EncodedLen(len(bMeta)))
	base64.StdEncoding.Encode(bMetaEncoded, bMeta)
	if noLineBreaks {
		bMetaEncoded = processLineBreaks(bMetaEncoded, []byte(""), lineBreakAfter)
	} else {
		bMetaEncoded = processLineBreaks(bMetaEncoded, lineBreakText, lineBreakAfter)
	}

	// Fix formatting on "LABEL" line if newly generated by apb init or galaxy init
	initialLabelRegexp, _ := regexp.Compile(`LABEL\s\"com\.redhat\.apb\.spec\"\=\\\n[^\"]`)
	indices := initialLabelRegexp.FindIndex(cMeta)
	if len(indices) > 0 {
		fixBytes := []byte("\"\"\n")
		insertAt := indices[1] - 1
		cMeta = append(cMeta[:insertAt], append(fixBytes, cMeta[insertAt:]...)...)
	}

	// Match the "LABEL" line up to the opening quote on the b64 blob
	labelRegexp, _ := regexp.Compile(`.*(\")?com\.redhat\.apb\.spec(\")?\=(\\)?\n?\"`)
	indices = labelRegexp.FindIndex(cMeta)
	if len(indices) == 0 {
		log.Errorf("Didn't find expected bundle label in container metadata file")
		return []byte{}
	}
	lineStartIndex, blobStartIndex := indices[0], indices[1]
	blobEndOffset := bytes.IndexByte(cMeta[blobStartIndex:], byte('"'))
	if blobEndOffset == -1 {
		log.Errorf("Didn't find end of bundle label in container metadata file")
		return []byte{}
	}
	blobEndIndex := blobStartIndex + blobEndOffset

	// Build new "LABEL" section for container metadata file
	bMetaSection := []byte{}
	bMetaSection = append(bMetaSection, cMeta[lineStartIndex:blobStartIndex-1]...)
	if !bytes.Contains(bMetaSection, lineBreakText) && !noLineBreaks {
		bMetaSection = append(bMetaSection, lineBreakText...)
	}
	bMetaSection = append(bMetaSection, byte('"'))
	bMetaSection = append(bMetaSection, bMetaEncoded...)

	// Build container metadata file output
	cMeta = append(cMeta[:lineStartIndex], append(bMetaSection, cMeta[blobEndIndex:]...)...)
	return cMeta
}

func processLineBreaks(text []byte, lineBreakText []byte, breakAfter int) []byte {
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
