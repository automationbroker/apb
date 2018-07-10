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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/automationbroker/apb/pkg/util"
	"github.com/automationbroker/bundle-lib/clients"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type bootstrapResponse struct {
	SpecCount  int `json:"spec_count"`
	ImageCount int `json:"image_count"`
}

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
	rootCmd.AddCommand(createHiddenCmd(brokerCatalogCmd, "running 'apb broker catalog'"))

	brokerCmd.AddCommand(brokerBootstrapCmd)
	rootCmd.AddCommand(createHiddenCmd(brokerBootstrapCmd, "running 'apb broker boostrap'"))
}

func listBrokerCatalog() {
	log.Debugf("func::listBrokerCatalog()")
	kube, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return
	}

	brokerRoute, err := getBrokerRoute(brokerName)
	if err != nil {
		log.Errorf("Failed to get broker route: %v", err)
	}

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
	if err != nil {
		log.Errorf("Failed fetch catalog: %v", err)
		return
	}
	printServices(services.Services)
	return
}

func bootstrapBroker() {
	kube, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return
	}

	// Get the broker route given brokerName
	brokerRoute, err := getBrokerRoute(brokerName)
	if err != nil {
		log.Errorf("Failed to get broker route: %v", err)
		return
	}

	// Create a new bootstrap request
	req, err := http.NewRequest("POST", fmt.Sprintf("%v/v2/bootstrap", brokerRoute), nil)
	if err != nil {
		log.Errorf("Failed to create request: %v", err)
		return
	}

	// Set Bearer token Auth
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", kube.ClientConfig.BearerToken))

	// Set TLS settings
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: cfg,
	}

	// Do bootstrap request
	fmt.Printf("Bootstrapping the broker at [%v/v2/bootstrap]. This may take up to a minute...\n", brokerRoute)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to get response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 504 {
		log.Errorf("Timed out waiting for broker bootstrap response.")
		fmt.Print("Try increasing the route timeout with:\n")
		fmt.Printf("oc annotate route asb -n %v --overwrite haproxy.router.openshift.io/timeout=60s\n", brokerName)
		return
	}

	if resp.StatusCode != 200 {
		log.Errorf("Failed to bootstrap the broker. Expected status 200, got: %v", resp.StatusCode)
		return
	}

	// Unmarshal response
	jsonBoot, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
	}

	bootResp := bootstrapResponse{}
	err = json.Unmarshal(jsonBoot, &bootResp)
	if err != nil {
		log.Errorf("Failed to unmarshal response: %v", err)
	}

	fmt.Printf("Successfully bootstrapped broker [%v]\n", brokerName)
	fmt.Printf("Broker loaded %v bundle specs from %v total images.\n", bootResp.SpecCount, bootResp.ImageCount)
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

func getBrokerRoute(brokerName string) (string, error) {
	var brokerRoute string
	ocp, err := clients.Openshift()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return "", err
	}

	// Attempt to get route of Automation Broker
	rc, err := ocp.Route().Routes(brokerName).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("Failed to list routes in namespace %v: %v", brokerName, err)
		return "", err
	}

	for _, route := range rc.Items {
		if route.Spec.To.Name == brokerName {
			brokerRoute = fmt.Sprintf("https://%v/%v", route.Spec.Host, brokerName)
			return brokerRoute, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Failed to find route for broker: %v", brokerName))
}
