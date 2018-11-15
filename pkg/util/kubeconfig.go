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

package util

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

// GetCurrentNamespace returns the current OpenShift namespace or an empty string
func GetCurrentNamespace(configPath string) string {
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		log.Errorf("Error loading kubeconfig from [%v]: %v", configPath, err)
		return ""
	}
	if len(strings.Split(config.CurrentContext, "/")) < 3 {
		log.Errorf("Did not find expected current-context string in kubeconfig.")
		return ""
	}
	return strings.Split(config.CurrentContext, "/")[0]
}

// GetKubeConfigPath returns a valid kubeconfig path
func GetKubeConfigPath(kcPath string) string {
	configPath := clientcmd.RecommendedHomeFile
	kubeconfigEnv := os.Getenv("KUBECONFIG")

	if len(kcPath) > 0 {
		configPath = kcPath
	} else if len(kubeconfigEnv) > 0 {
		configPath = kubeconfigEnv
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Errorf("Error loading kubeconfig from [%v]: %v (Try setting the kubeconfig path with -k, --kubeconfig)", configPath, err)
		os.Exit(1)
	}

	return configPath
}
