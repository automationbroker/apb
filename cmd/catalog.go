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
	//"encoding/json"
	"crypto/tls"
	"fmt"
	"github.com/automationbroker/bundle-lib/clients"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Interact with OpenShift Service Catalog",
	Long:  `Force the OpenShift Service Catalog to relist its group of APB specs`,
}

var catalogRelistCmd = &cobra.Command{
	Use:   "relist",
	Short: "relist service catalog",
	Long:  `Force a relist of the OpenShift Service Catalog`,
	Run: func(cmd *cobra.Command, args []string) {
		relistCatalog()
	},
}

func init() {
	rootCmd.AddCommand(catalogCmd)
	// Catalog Relist Flags
	catalogCmd.AddCommand(catalogRelistCmd)
}

func relistCatalog() {
	log.Debug("relistCatalog called")
	kube, err := clients.Kubernetes()
	if err != nil {
		log.Errorf("Failed to connect to cluster: %v", err)
		return
	}
	host := kube.ClientConfig.Host
	brokerUrl := fmt.Sprintf("%v/apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/%v", host, "ansible-service-broker")
	req, err := http.NewRequest("GET", brokerUrl, nil)
	if err != nil {
		log.Errorf("Failed to create relist request: %v", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", kube.ClientConfig.BearerToken))
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: cfg,
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Failed to get relist response: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Errorf("Failed to get relist response. Expected status 200, got: %v", resp.StatusCode)
		return
	}
	fmt.Printf("%#v\n", resp.Body)
	return
}
