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
package config

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Defaults stores command defaults
var Defaults *viper.Viper
var LoadedDefaults DefaultSettings

// Registries stores APB registry and spec data
var Registries *viper.Viper

// InitJSONConfig will load (or create if needed) a JSON config at ~/home/.apb/configName.json or configDir/configName.json
func InitJSONConfig(configDir string, configName string) (config *viper.Viper, isNewConfig bool) {
	var configPath string

	viperConfig := viper.New()
	viperConfig.SetConfigType("json")
	if configDir != "" {
		viperConfig.AddConfigPath(configDir)
		configPath = configDir
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		configPath = filepath.Join(home, ".apb")
	}
	viperConfig.AddConfigPath(configPath)
	viperConfig.SetConfigName(configName)
	filePath := configPath + fmt.Sprintf("/%s.json", configName)
	if err := viperConfig.ReadInConfig(); err != nil {
		isNewConfig = true
		log.Warningf("Didn't find config file %s, creating.", filePath)
		os.MkdirAll(configPath, 0755)
		file, err := os.Create(filePath)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		file.WriteString("{}")
	}
	if err := viperConfig.ReadInConfig(); err != nil {
		log.Error("Can't read config: ", err)
		os.Exit(1)
	}
	return viperConfig, isNewConfig
}

// UpdateCachedRegistries saves the contents of regList to a configuration file
func UpdateCachedRegistries(viperConfig *viper.Viper, regList []Registry) error {
	viperConfig.Set("Registries", regList)
	viperConfig.WriteConfig()
	return nil
}

// UpdateCachedDefaults saves the contents of defaults to a configuration file
func UpdateCachedDefaults(viperConfig *viper.Viper, defaults *DefaultSettings) error {
	viperConfig.Set("Defaults", defaults)
	viperConfig.WriteConfig()
	return nil
}

// InitialDefaultSettings provides a DefaultSettings struct with sane values for interaction with the Automation Broker
func InitialDefaultSettings() *DefaultSettings {
	return &DefaultSettings{
		BrokerNamespace:          "openshift-automation-service-broker",
		BrokerResourceURL:        "/apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/",
		BrokerRouteName:          "openshift-automation-service-broker",
		ClusterServiceBrokerName: "openshift-automation-service-broker",
	}
}

// LoadDefaultSettings loads default settings from disk into a config.LoadedDefaults for later use
func LoadDefaultSettings(viperConfig *viper.Viper, defaults *DefaultSettings) {
	viperConfig.UnmarshalKey("Defaults", defaults)
}
