package util

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

func GetCurrentNamespace(configPath string) string {
	if configPath == "" {
		configPath = clientcmd.RecommendedHomeFile
	}
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

func GetCurrentCluster(configPath string) string {
	if configPath == "" {
		configPath = clientcmd.RecommendedHomeFile
	}
	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		log.Errorf("Error loading kubeconfig from [%v]: %v", configPath, err)
		return ""
	}
	if len(strings.Split(config.CurrentContext, "/")) < 3 {
		log.Errorf("Did not find expected current-context string in kubeconfig.")
		return ""
	}
	return strings.Split(config.CurrentContext, "/")[1]
}
