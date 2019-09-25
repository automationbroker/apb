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
	"os"

	"github.com/automationbroker/apb/pkg/config"
	"github.com/automationbroker/apb/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Verbose controls the logging level, when enabled will set level to debug
var Verbose bool

var cfgDir string

var kubeConfig string

var rootCmd = &cobra.Command{
	Use:   "apb",
	Short: "Tool for working with Ansible Playbook Bundles",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&cfgDir, "config", "", "configuration directory (default is $HOME/.apb)")
	rootCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "k", "", "Path to kubeconfig to use (default is $HOME/.kube/config)")
}

func initConfig() {
	var isNewDefaultsConfig bool

	// Load or create registries.json
	config.Registries, _ = config.InitJSONConfig(cfgDir, "registries")
	// Load or create defaults.json
	config.Defaults, isNewDefaultsConfig = config.InitJSONConfig(cfgDir, "defaults")
	if isNewDefaultsConfig {
		config.UpdateCachedDefaults(config.Defaults, config.InitialDefaultSettings())
	}
	config.LoadDefaultSettings(config.Defaults, &config.LoadedDefaults)

	kubeConfig = util.GetKubeConfigPath(kubeConfig)
	os.Setenv("KUBECONFIG", kubeConfig)
}

// Execute invokes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
