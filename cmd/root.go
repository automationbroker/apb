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

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Verbose controls the logging level, when enabled will set level to debug
var Verbose bool

// CfgFile is the configuration file
var CfgFile string

// Config sbcli configuration
type Config struct {
	Registries []registries.Config
	Specs      []*bundle.Spec
}

var rootCmd = &cobra.Command{
	Use:   "sbcli",
	Short: "sbcli is a tool to manage ServiceBundle images",
	Long: `ServiceBundles are images that represent lifecycle components
in that they contain all of the orchestration logic to manage
an application through out it's lifecycle, i.e. install, uninstall,
bind, unbind, etc.  ServiceBundles are intended to be invoked and run
as a short job to execute the intended work, example I want to deploy a
postgres database to my kubernetes cluster.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func init() {
	log.SetLevel(log.InfoLevel)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "configuration file (default is $HOME/.sbcli)")
}

func initConfig() {
	viper.SetConfigType("json")
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".sbcli")
		filePath := home + "/.sbcli.json"
		if err := viper.ReadInConfig(); err != nil {
			log.Warning("Didn't find config file, creating one.")
			file, err := os.Create(filePath)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			file.WriteString("{}")
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Error("Can't read config: ", err)
		os.Exit(1)
	}
}

// Execute invokes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
