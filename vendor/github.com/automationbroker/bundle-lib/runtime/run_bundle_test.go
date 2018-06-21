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

package runtime

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/automationbroker/bundle-lib/clients"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	fakerest "k8s.io/client-go/rest/fake"
)

func TestDefaultRunBundle(t *testing.T) {
	restClient := &fakerest.RESTClient{
		Resp: &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"major":"3", "minor": "2"}`))),
		},
		NegotiatedSerializer: scheme.Codecs,
	}
	var optionalFalse bool
	cases := []struct {
		name        string
		exContext   ExecutionContext
		expectedEX  ExecutionContext
		expectedPod *v1.Pod
		client      *fake.Clientset
		shouldErr   bool
		validatePod func(t *testing.T, pod *v1.Pod)
	}{
		{
			name: "run bundle successfully",
			exContext: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "Always",
			},
			expectedEX: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "Always",
			},
			expectedPod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bundle-test",
					Namespace: "test-bundle-test",
					Labels:    nil,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "apb",
							Image: "new-image",
							Args: []string{
								"provision",
								"--extra-vars",
								`{"apb": "test"}`,
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								v1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							ImagePullPolicy: v1.PullAlways,
							VolumeMounts:    []v1.VolumeMount{},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "svc-acct-bundle-test",
					Volumes:            []v1.Volume{},
				},
			},
			client: fake.NewSimpleClientset(),
		},
		{
			name: "run bundle successfully with a state mounted",
			exContext: ExecutionContext{
				BundleName:    "bundle-test-state",
				Account:       "svc-acct-bundle-test",
				Action:        "provision",
				Location:      "test-bundle-test",
				Targets:       []string{"target-bundle-test"},
				Secrets:       []string{},
				ExtraVars:     `{"apb": "test"}`,
				Image:         "new-image",
				Policy:        "Never",
				StateName:     "bundle-test-state",
				StateLocation: defaultMountLocation,
			},
			expectedEX: ExecutionContext{
				BundleName:    "bundle-test-state",
				Account:       "svc-acct-bundle-test",
				Action:        "provision",
				Location:      "test-bundle-test",
				Targets:       []string{"target-bundle-test"},
				Secrets:       []string{},
				ExtraVars:     `{"apb": "test"}`,
				Image:         "new-image",
				Policy:        "Never",
				StateName:     "bundle-test-state",
				StateLocation: defaultMountLocation,
			},
			validatePod: func(t *testing.T, pod *v1.Pod) {
				if pod.Name != "bundle-test-state" {
					t.Fatalf("expected pod name to be %s but was %s", "bundle-test-state", pod.Name)
				}
				if len(pod.Spec.Containers) != 1 {
					t.Fatalf("expected to have 1 container but got %v ", len(pod.Spec.Containers))
				}
				container := pod.Spec.Containers[0]

				var (
					podName, podNameSpace, bundleState bool
					bundleStateLocation                string
				)
				if len(container.Env) != 3 {
					t.Fatalf("expected 3 envVars but there was %d", len(container.Env))
				}
				for _, e := range container.Env {
					if e.Name == "POD_NAME" {
						podName = true
					}
					if e.Name == "POD_NAMESPACE" {
						podNameSpace = true
					}
					if e.Name == "BUNDLE_STATE_LOCATION" {
						bundleStateLocation = e.Value
						bundleState = true
					}
				}

				if !podName || !podNameSpace || !bundleState {
					t.Fatalf("expected to find POD_NAME, POD_NAMESPACE, BUNDLE_STATE_LOCATION env vars")
				}
				if len(container.VolumeMounts) != 1 {
					t.Fatalf("expected 1 volume mount but got %v ", len(container.VolumeMounts))
				}
				if container.VolumeMounts[0].MountPath != bundleStateLocation {
					t.Fatalf("expected the mount path to be %s but was %s ", bundleStateLocation, container.VolumeMounts[0].MountPath)
				}
				if len(pod.Spec.Volumes) != 1 {
					t.Fatalf("expected 1 volume but got %v ", len(pod.Spec.Volumes))
				}
				if pod.Spec.Volumes[0].ConfigMap == nil {
					t.Fatalf("expected the volume to be a configmap but was nil")
				}
				if pod.Spec.Volumes[0].ConfigMap.Name != pod.Name {
					t.Fatalf("expected the configmap name to match the pod name but it was %s ", pod.Spec.Volumes[0].ConfigMap.Name)
				}

			},
			client: fake.NewSimpleClientset(),
		},
		{
			name: "run bundle successfully with a secret mounted",
			exContext: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{"test-secret"},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "Never",
			},
			expectedEX: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{"test-secret"},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "Never",
			},
			expectedPod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bundle-test",
					Namespace: "test-bundle-test",
					Labels:    nil,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "apb",
							Image: "new-image",
							Args: []string{
								"provision",
								"--extra-vars",
								`{"apb": "test"}`,
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								v1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							ImagePullPolicy: v1.PullNever,
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "apb-test-secret",
									MountPath: "/etc/apb-secrets/" + "apb-test-secret",
									ReadOnly:  true,
								},
							},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "svc-acct-bundle-test",
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "apb-test-secret",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "test-secret",
									Optional:   &optionalFalse,
								},
							},
						},
					},
				},
			},
			client: fake.NewSimpleClientset(),
		},
		{
			name: "run bundle successfully with proxy config",
			exContext: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "IfNotPresent",
				ProxyConfig: &ProxyConfig{
					HTTPProxy:  "http://foo.com",
					HTTPSProxy: "https://foo.com",
					NoProxy:    "*.local",
				},
			},
			expectedEX: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "IfNotPresent",
				ProxyConfig: &ProxyConfig{
					HTTPProxy:  "http://foo.com",
					HTTPSProxy: "https://foo.com",
					NoProxy:    "*.local",
				},
			},
			expectedPod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bundle-test",
					Namespace: "test-bundle-test",
					Labels:    nil,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "apb",
							Image: "new-image",
							Args: []string{
								"provision",
								"--extra-vars",
								`{"apb": "test"}`,
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								v1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								v1.EnvVar{
									Name:  httpProxyEnvVar,
									Value: "http://foo.com",
								},
								v1.EnvVar{
									Name:  httpsProxyEnvVar,
									Value: "https://foo.com",
								},
								v1.EnvVar{
									Name:  noProxyEnvVar,
									Value: "*.local",
								},
								v1.EnvVar{
									Name:  strings.ToLower(httpProxyEnvVar),
									Value: "http://foo.com",
								},
								v1.EnvVar{
									Name:  strings.ToLower(httpsProxyEnvVar),
									Value: "https://foo.com",
								},
								v1.EnvVar{
									Name:  strings.ToLower(noProxyEnvVar),
									Value: "*.local",
								},
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							VolumeMounts:    []v1.VolumeMount{},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "svc-acct-bundle-test",
					Volumes:            []v1.Volume{},
				},
			},
			client: fake.NewSimpleClientset(),
		},
		{
			name:      "invalid k8scli",
			client:    nil,
			shouldErr: true,
		},
		{
			name: "invalid pull policy",
			exContext: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "AlwaysNotAnything",
			},
			client:    fake.NewSimpleClientset(),
			shouldErr: true,
		},
		{
			name: "pod already exists error",
			exContext: ExecutionContext{
				BundleName: "bundle-test",
				Account:    "svc-acct-bundle-test",
				Action:     "provision",
				Location:   "test-bundle-test",
				Targets:    []string{"target-bundle-test"},
				Secrets:    []string{},
				ExtraVars:  `{"apb": "test"}`,
				Image:      "new-image",
				Policy:     "Always",
			},
			client: fake.NewSimpleClientset(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bundle-test",
					Namespace: "test-bundle-test",
					Labels:    nil,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "apb",
							Image: "new-image",
							Args: []string{
								"provision",
								"--extra-vars",
								`{"apb": "test"}`,
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								v1.EnvVar{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							ImagePullPolicy: v1.PullAlways,
							VolumeMounts:    []v1.VolumeMount{},
						},
					},
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "svc-acct-bundle-test",
					Volumes:            []v1.Volume{},
				},
			}),
			shouldErr: true,
		},
	}
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if nil != tc.client {
				k.Client = &fakeClientSet{
					tc.client,
					restClient,
				}
				NewRuntime(Configuration{})
			}
			actualEXContext, err := defaultRunBundle(tc.exContext)
			if err != nil {
				if !tc.shouldErr {
					t.Fatalf("unknown error: %v", err)
				}
				return
			}
			if tc.shouldErr {
				t.Fatalf("expected error but did not recieve one")
			}
			//Get the pod that should have been created.
			actualPod, err := k.Client.CoreV1().Pods(actualEXContext.Location).Get(actualEXContext.BundleName, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("retrieval of the pod failed - %v", err)
			}
			if tc.validatePod != nil {
				tc.validatePod(t, actualPod)
			} else {
				if !reflect.DeepEqual(actualPod, tc.expectedPod) {
					if len(actualPod.Spec.Volumes) > 0 {
						fmt.Printf("\ngot: %#v\nexp: %#v", actualPod.Spec.Volumes[0].Secret, tc.expectedPod.Spec.Volumes[0].Secret)
					}
					t.Fatalf("Unexpected pod\n\nGot: %#+v\nExpected: %#+v\n", actualPod, tc.expectedPod)
				}
			}
			if !reflect.DeepEqual(actualEXContext, tc.expectedEX) {
				t.Fatalf("Unexpected ex context\n\nGot: %#+v\nExpected: %#+v\n", actualEXContext, tc.expectedEX)
			}
		})
	}
}

func TestDefaultCopySecretsToNamespace(t *testing.T) {
	cases := []struct {
		name           string
		ec             ExecutionContext
		cn             string
		secrets        []string
		shouldError    bool
		client         *fake.Clientset
		expectedSecret *v1.Secret
	}{
		{
			name: "copy no secrets",
			ec: ExecutionContext{
				Location: "test",
			},
			cn:      "cluster-test",
			secrets: []string{},
			client:  fake.NewSimpleClientset(),
		},
		{
			name: "copy secret",
			ec: ExecutionContext{
				Location: "test",
			},
			cn:      "cluster-test",
			secrets: []string{"test-secret"},
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "cluster-test",
					Labels: map[string]string{
						"label": "value",
					},
					Annotations: map[string]string{
						"annotation": "value",
					},
				},
				StringData: map[string]string{"hello": "world"},
			}),
			expectedSecret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "test",
					Labels: map[string]string{
						"label": "value",
					},
					Annotations: map[string]string{
						"annotation": "value",
					},
				},
				StringData: map[string]string{"hello": "world"},
			},
		},
		{
			name: "secret not found",
			ec: ExecutionContext{
				Location: "test",
			},
			cn:          "cluster-test",
			secrets:     []string{"test-secret"},
			client:      fake.NewSimpleClientset(),
			shouldError: true,
		},
		{
			name: "secret already copied error",
			ec: ExecutionContext{
				Location: "test",
			},
			cn:      "cluster-test",
			secrets: []string{"test-secret"},
			client: fake.NewSimpleClientset(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "cluster-test",
					Labels: map[string]string{
						"label": "value",
					},
					Annotations: map[string]string{
						"annotation": "value",
					},
				},
				StringData: map[string]string{"hello": "world"},
			},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "test",
						Labels: map[string]string{
							"label": "value",
						},
						Annotations: map[string]string{
							"annotation": "value",
						},
					},
					StringData: map[string]string{"hello": "world"},
				}),
			shouldError: true,
		},
	}
	k, err := clients.Kubernetes()
	if err != nil {
		t.Fail()
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k.Client = tc.client
			err := defaultCopySecretsToNamespace(tc.ec, tc.cn, tc.secrets)
			if err != nil {
				if !tc.shouldError {
					t.Fatalf("unknown error occurred")
				}
				return
			}
			// If nothing was supposed to be created then nothing should be in the namespace.
			secrets, err := k.Client.CoreV1().Secrets(tc.ec.Location).List(metav1.ListOptions{})
			if tc.expectedSecret == nil && len(secrets.Items) == 0 {
				return
			}
			if tc.expectedSecret != nil && len(secrets.Items) == 0 {
				t.Fatalf("secret was not copied")
			}
			if !reflect.DeepEqual(tc.expectedSecret, &secrets.Items[0]) {
				t.Fatalf("unexpected secret:\nGot: %#v\nExp: %v", secrets.Items[0], tc.expectedSecret)
			}
		})
	}
}
