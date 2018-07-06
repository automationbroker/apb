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
	return strings.Split(config.CurrentContext, "/")[0]
}
