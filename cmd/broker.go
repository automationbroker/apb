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
	"strconv"

	"github.com/automationbroker/bundle-lib/clients"
	"github.com/automationbroker/sbcli/pkg/util"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Long:  `Refresh list of bootstrapped service bundles in Automation Broker catalog`,
	Run: func(cmd *cobra.Command, args []string) {
		bootstrapBroker()
	},
}

func init() {
	brokerCmd.PersistentFlags().StringVarP(&brokerName, "name", "n", "openshift-automation-service-broker", "Name of Automation Broker instance")
	rootCmd.AddCommand(brokerCmd)

	brokerCmd.AddCommand(brokerCatalogCmd)
	brokerCmd.AddCommand(brokerBootstrapCmd)
}

func listBrokerCatalog() {
	log.Debugf("func::listBrokerCatalog()")
	var brokerRoute string
	kube, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return
	}
	ocp, err := clients.Openshift()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return
	}

	// Attempt to get route of Automation Broker
	rc, err := ocp.Route().Routes(brokerName).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("Failed to list routes in namespace %v: %v", brokerName, err)
	}

	for _, route := range rc.Items {
		if route.Spec.To.Name == brokerName {
			brokerRoute = route.Spec.Host
			break
		}
	}
	if brokerRoute == "" {
		log.Errorf("Failed to find broker route.")
		return
	}

	brokerRoute = fmt.Sprintf("https://%v/%v", brokerRoute, brokerName)

	osbConf := &osb.ClientConfiguration{
		Name:                "automation-broker",
		URL:                 brokerRoute,
		APIVersion:          osb.LatestAPIVersion(),
		TimeoutSeconds:      60,
		EnableAlphaFeatures: false,
		Insecure:            true,
		AuthConfig: &osb.AuthConfig{
			BearerConfig: &osb.BearerConfig{
				Token: kube.ClientConfig.BearerToken,
			},
		},
	}
	osbClient, err := osb.NewClient(osbConf)
	if err != nil {
		log.Errorf("Failed to make osb client: %v", err)
		return
	}

	services, err := osbClient.GetCatalog()
	printServices(services.Services)
	return
}

func bootstrapBroker() {
	return
}

func printServices(services []osb.Service) {
	colName := &util.TableColumn{Header: "NAME"}
	colID := &util.TableColumn{Header: "ID"}
	colBind := &util.TableColumn{Header: "BINDABLE"}

	for _, s := range services {
		colName.Data = append(colName.Data, s.Name)
		colID.Data = append(colID.Data, s.ID)
		colBind.Data = append(colBind.Data, strconv.FormatBool(s.Bindable))
	}

	tableToPrint := []*util.TableColumn{colName, colID, colBind}
	util.PrintTable(tableToPrint)
}
