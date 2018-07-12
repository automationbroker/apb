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
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set tool defaults",
	Long:  `Runs an interactive prompt to configure defaults for the 'apb' tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		gatherConfig()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func gatherConfig() {
	defaultSettings := &config.DefaultSettings{
		BrokerNamespace:          getUserInput("Broker namespace", "openshift-automation-service-broker"),
		BrokerResourceURL:        getUserInput("Broker resource URL", "/apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/"),
		BrokerRouteName:          getUserInput("Broker route name", "openshift-automation-service-broker"),
		ClusterServiceBrokerName: getUserInput("clusterservicebroker name", "openshift-automation-service-broker"),
	}
	config.UpdateCachedDefaults(defaultSettings)
}

func getUserInput(prompt string, defaultValue string) string {
	var userInput string
	fmt.Printf("%s [default: %s]: ", prompt, defaultValue)
	fmt.Scanln(&userInput)
	if userInput == "" {
		return defaultValue
	}
	return userInput
}
