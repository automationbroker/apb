package cmd

import (
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
	Short: "Interact with ServiceBundle images",
	Long:  `Interact with and list ServiceBundles from configured registry adapters`,
}

var bundleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ServiceBundle images",
	Long:  `List ServiceBundles from a registry adapter`,
	Run: func(cmd *cobra.Command, args []string) {
		listImages()
	},
}

func init() {
	rootCmd.AddCommand(bundleCmd)
	bundleListCmd.Flags().BoolVarP(&Refresh, "refresh", "r", false, "refresh list of specs")
	bundleCmd.AddCommand(bundleListCmd)
}

func updateCachedList(specs []*bundle.Spec) error {
	viper.Set("Specs", specs)
	viper.WriteConfig()
	return nil
}

func getImages() ([]*bundle.Spec, error) {
	var regConfigList []registries.Config
	var regList []registries.Registry
	var specList []*bundle.Spec
	err := viper.UnmarshalKey("Registries", &regConfigList)
	if err != nil {
		log.Error("Error unmarshalling config: ", err)
		return nil, err
	}

	authNamespace := ""
	for _, config := range regConfigList {
		registry, err := registries.NewRegistry(config, authNamespace)
		if err != nil {
			log.Error("Error from creating a NewRegistry")
			log.Error(err)
			return nil, err
		}
		regList = append(regList, registry)
	}
	for _, reg := range regList {
		specs, count, err := reg.LoadSpecs()
		if err != nil {
			log.Errorf("registry: %v was unable to complete bootstrap - %v",
				reg.RegistryName(), err)
			return nil, err
		}
		log.Infof("Registry %v has %d bundles available from %d images scanned", reg.RegistryName(), len(specs), count)
		specList = append(specList, specs...)
	}

	return specList, nil
}

func listImages() {
	var specs []*bundle.Spec
	err := viper.UnmarshalKey("Specs", &specs)
	if err != nil {
		log.Error("Error unmarshalling config: ", err)
		return
	}
	if len(specs) > 0 && Refresh == false {
		log.Println("Found specs already in config")
		for _, s := range specs {
			log.Printf("%v - %v\n", s.FQName, s.Image)
		}
		return
	}

	specs, err = getImages()
	if err != nil {
		log.Error("Error getting images")
		return
	}
	log.Printf("specs: %v\n", specs)
	err = updateCachedList(specs)
	if err != nil {
		log.Error("Error updating cache")
		return
	}

	for _, s := range specs {
		log.Printf("%v - %v\n", s.FQName, s.Image)
	}
}
