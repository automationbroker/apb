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

package adapters

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/clients"
	v1image "github.com/openshift/api/image/v1"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	errRuntimeNotFound = errors.New("runtime not found")
)

type imageLabel struct {
	Spec    string `json:"com.redhat.apb.spec"`
	Runtime string `json:"com.redhat.apb.runtime"`
}

type containerConfig struct {
	Labels imageLabel `json:"Labels"`
}

type imageMetadata struct {
	ContainerConfig containerConfig `json:"ContainerConfig"`
}

const localOpenShiftName = "openshift-registry"

// LocalOpenShiftAdapter - Docker Hub Adapter
type LocalOpenShiftAdapter struct {
	Config Configuration
}

// RegistryName - Retrieve the registry name
func (r LocalOpenShiftAdapter) RegistryName() string {
	return localOpenShiftName
}

// GetImageNames - retrieve the images
func (r LocalOpenShiftAdapter) GetImageNames() ([]string, error) {
	log.Debug("LocalOpenShiftAdapter::GetImageNames")
	log.Debugf("BundleSpecLabel: %s", BundleSpecLabel)

	openshiftClient, err := clients.Openshift()
	if err != nil {
		log.Errorf("Failed to instantiate OpenShift client")
		return nil, err
	}

	images, err := openshiftClient.ListRegistryImages()
	if err != nil {
		log.Errorf("Failed to load registry images")
		return nil, err
	}

	imageList := []string{}
	for _, image := range images.Items {
		imageList = append(imageList, strings.Split(image.DockerImageManifest, "@")[0])
	}

	return imageList, nil
}

// FetchSpecs - retrieve the spec for the image names.
func (r LocalOpenShiftAdapter) FetchSpecs(imageNames []string) ([]*bundle.Spec, error) {
	log.Debug("LocalOpenShiftAdapter::FetchSpecs")
	specList := []*bundle.Spec{}
	registryIP, err := r.getServiceIP("docker-registry", "default")
	if err != nil {
		log.Errorf("Failed get docker-registry service information.")
		return nil, err
	}

	openshiftClient, err := clients.Openshift()
	if err != nil {
		log.Errorf("Failed to instantiate OpenShift client.")
		return nil, err
	}

	listImages, err := openshiftClient.ListRegistryImages()
	if err != nil {
		log.Errorf("Failed to load registry images")
		return nil, err
	}

	for _, image := range listImages.Items {
		n := strings.Split(image.DockerImageManifest, "@")[0]
		for _, providedImage := range imageNames {
			if providedImage == n {
				spec, err := r.loadSpec(image)
				if err != nil {
					log.Errorf("Failed to load image spec")
					continue
				}
				if strings.HasPrefix(n, registryIP) == false {
					log.Debugf("Image does not have a registry IP as prefix. This might cause problems but not erroring out.")
				}
				if r.Config.Namespaces == nil {
					log.Debugf("Namespace not set. Assuming `openshift`")
					r.Config.Namespaces = append(r.Config.Namespaces, "openshift")
				}
				spec.Image = n
				nsList := strings.Split(n, "/")
				var namespace string
				if len(nsList) == 0 {
					log.Errorf("Image [%v] is not in the proper format. Erroring.", n)
					continue
				} else if len(nsList) < 3 {
					// Image does not have any registry prefix. May be a product of S2I
					// Expecting openshift/foo-bundle
					namespace = nsList[0]
				} else {
					// Expecting format: 172.30.1.1:5000/openshift/foo-bundle
					namespace = nsList[1]
				}
				for _, ns := range r.Config.Namespaces {
					// logging to warn users about the potential bug if
					// the svc-acct does not have access to the namespace.
					if ns != "openshift" {
						log.Warningf("You may not be able to load provision images from the namespace: %v.\n"+
							"You should make sure that the namespace has given the permissions for the "+
							"system:authenticated group.", ns)
					}
					if ns == namespace {
						log.Debugf("Image [%v] is in configured namespace [%v]. Adding to SpecList.", n, ns)
						specList = append(specList, spec)
					}
				}
			}
		}
	}

	return specList, nil
}

func (r LocalOpenShiftAdapter) loadSpec(image v1image.Image) (*bundle.Spec, error) {
	log.Debug("LocalOpenShiftAdapter::LoadSpec")
	b, err := image.DockerImageMetadata.MarshalJSON()
	if err != nil {
		log.Errorf("unable to get json docker image metadata: %v", err)
		return nil, err
	}
	i := imageMetadata{}
	err = json.Unmarshal(b, &i)
	if err != nil {
		log.Errorf("unable to get unmarshal json docker image metadata: %v", err)
		return nil, err
	}
	spec := &bundle.Spec{}

	err = yaml.Unmarshal([]byte(i.ContainerConfig.Labels.Spec), spec)
	if err != nil {
		log.Errorf("Something went wrong loading decoded spec yaml, %s", err)
		return nil, err
	}
	spec.Runtime, err = getAPBRuntimeVersion(i.ContainerConfig.Labels.Runtime)
	if err != nil {
		log.Errorf("Failed to parse image runtime version")
		return nil, errRuntimeNotFound
	}
	return spec, nil
}

func (r LocalOpenShiftAdapter) getServiceIP(service string, namespace string) (string, error) {
	k8s, err := clients.Kubernetes()
	if err != nil {
		return "", err
	}

	serviceData, err := k8s.Client.CoreV1().Services(namespace).Get(service, meta_v1.GetOptions{})
	if err != nil {
		log.Warningf("Unable to load service '%s' from namespace '%s'", service, namespace)
		return "", err
	}
	log.Debugf("Found service with name %v", service)

	return serviceData.Spec.ClusterIP, nil
}
