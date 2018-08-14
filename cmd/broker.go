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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"net/http"
	"strconv"
	"strings"

	"github.com/automationbroker/apb/pkg/config"

	"github.com/automationbroker/apb/pkg/util"
	"github.com/automationbroker/bundle-lib/clients"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var brokerName string
var catalogOutputFormat string

type ServiceEncoder interface {
	Encode(interface{}) error
}

var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "Interact with Automation Broker",
	Long:  `List or Bootstrap APBs on an Automation Broker instance`,
}

var brokerCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "List available APBs in Automation Broker catalog",
	Long:  `Fetch list of APBs in Automation Broker catalog`,
	Run: func(cmd *cobra.Command, args []string) {
		pruneOutputFormat(catalogOutputFormat)
		listBrokerCatalog(config.LoadedDefaults.BrokerRouteName, config.LoadedDefaults.BrokerNamespace)
	},
}

var brokerBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap an Automation Broker instance",
	Long:  `Refresh list of bootstrapped APBs in Automation Broker catalog`,
	Run: func(cmd *cobra.Command, args []string) {
		bootstrapBroker(config.LoadedDefaults.BrokerRouteName, config.LoadedDefaults.BrokerNamespace)
	},
}

func init() {
	brokerCmd.PersistentFlags().StringVarP(&brokerName, "name", "n", "", "Name of Automation Broker instance")
	rootCmd.AddCommand(brokerCmd)

	brokerCatalogCmd.Flags().StringVarP(&catalogOutputFormat, "output", "o", "", "Display broker catalog output in different format (json or yaml)")
	brokerCmd.AddCommand(brokerCatalogCmd)

	brokerCmd.AddCommand(brokerBootstrapCmd)
	rootCmd.AddCommand(createHiddenCmd(brokerBootstrapCmd, "running 'apb broker bootstrap'"))
}

func pruneOutputFormat(format string) {
	if format != "yaml" && format != "json" && format != "" {
		log.Warnf("Did not recognize --output argument [%v], printing as table", format)
	}
	return
}

func listBrokerCatalog(brokerRouteName string, brokerNamespace string) {
	log.Debugf("func::listBrokerCatalog()")
	// Override configured values if user provides brokerName as cmd arg
	if brokerName != "" {
		brokerRouteName = brokerName
		brokerNamespace = brokerName
	}
	kube, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return
	}

	// Check for user with valid bearer token
	if kube.ClientConfig.BearerToken == "" {
		handleBearerTokenErr()
		return
	}

	brokerRoute, err := getBrokerRoute(brokerRouteName, brokerNamespace)
	if err != nil {
		log.Errorf("Failed to get broker route: %v", err)
		if strings.Contains(err.Error(), "cannot list routes") {
			handleResourceInaccessibleErr("routes", brokerRouteName, false)
		}
		return
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

	printServices(services.Services, catalogOutputFormat)
	return
}

func bootstrapBroker(brokerRouteName string, brokerNamespace string) {
	// Override configured values if user provides brokerName as cmd arg
	if brokerName != "" {
		brokerRouteName = brokerName
		brokerNamespace = brokerName
	}
	kube, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return
	}

	// Check for user with valid bearer token
	if kube.ClientConfig.BearerToken == "" {
		handleBearerTokenErr()
		return
	}

	// Get the broker route given brokerName
	brokerRoute, err := getBrokerRoute(brokerRouteName, brokerNamespace)
	if err != nil {
		log.Errorf("Failed to get broker route: %v", err)
		if strings.Contains(err.Error(), "cannot list routes") {
			handleResourceInaccessibleErr("routes", brokerRouteName, false)
		}
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
	fmt.Printf("Bootstrapping the broker at [%v/v2/bootstrap].\n", brokerRoute)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to get response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 504 {
		log.Errorf("Timed out waiting for broker bootstrap response.")
		fmt.Print("Try increasing the route timeout with:\n")
		fmt.Printf("oc annotate route asb -n %v --overwrite haproxy.router.openshift.io/timeout=60s\n", brokerRouteName)
		return
	}

	if resp.StatusCode != 200 {
		log.Errorf("Failed to bootstrap the broker. Expected status 200, got: %v", resp.StatusCode)
		return
	}

	fmt.Printf("Successfully started bootstrap job for broker [%v]\n", brokerRouteName)
	return
}

func printServices(services []osb.Service, format string) {
	var encoder ServiceEncoder
	buffer := new(bytes.Buffer)

	switch format {
	case "json":
		enc := json.NewEncoder(buffer)
		enc.SetIndent("", "    ")
		encoder = enc
	case "yaml":
		encoder = yaml.NewEncoder(buffer)
	default:
		printServicesAsTable(services)
		return
	}
	err := encoder.Encode(services)
	if err != nil {
		log.Errorf("Failed to encode services data: [%v]", err)
		return
	}
	fmt.Printf("%v", buffer.String())
	return
}

func printServicesAsTable(services []osb.Service) {
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

func getBrokerRoute(brokerRouteName string, brokerNamespace string) (string, error) {
	// Override configured values if user provides brokerName as cmd arg
	if brokerName != "" {
		brokerRouteName = brokerName
		brokerNamespace = brokerName
	}
	var brokerRoute string
	ocp, err := clients.Openshift()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return "", err
	}

	// Attempt to get route of Automation Broker
	rc, err := ocp.Route().Routes(brokerNamespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, route := range rc.Items {
		if route.Spec.To.Name == brokerRouteName {
			brokerRoute = fmt.Sprintf("https://%v/%v", route.Spec.Host, brokerRouteName)
			return brokerRoute, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Failed to find route with name [%v] in namespace [%v]", brokerRouteName, brokerNamespace))
}
