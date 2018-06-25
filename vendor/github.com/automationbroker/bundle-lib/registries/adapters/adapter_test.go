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
	"testing"

	"fmt"

	"github.com/automationbroker/bundle-lib/bundle"
	ft "github.com/stretchr/testify/assert"
)

func TestSpecLabel(t *testing.T) {
	ft.Equal(t, BundleSpecLabel, "com.redhat.apb.spec", "spec label does not match dockerhub")
}

func TestImageToSpec(t *testing.T) {
	var testApbSpec = "dmVyc2lvbjogMS4wCm5hbWU6IG1lZGlhd2lraS1hcGIKZGVzY3JpcHRpb246IE1lZGlhd2lraSBhcGIgaW1wbGVtZW50YXRpb24KYmluZGFibGU6IEZhbHNlCmFzeW5jOiBvcHRpb25hbAptZXRhZGF0YToKICBkb2N1bWVudGF0aW9uVXJsOiBodHRwczovL3d3dy5tZWRpYXdpa2kub3JnL3dpa2kvRG9jdW1lbnRhdGlvbgogIGxvbmdEZXNjcmlwdGlvbjogQW4gYXBiIHRoYXQgZGVwbG95cyBNZWRpYXdpa2kgMS4yMwogIGRlcGVuZGVuY2llczogWydkb2NrZXIuaW8vYW5zaWJsZXBsYXlib29rYnVuZGxlL21lZGlhd2lraTEyMzpsYXRlc3QnXQogIGRpc3BsYXlOYW1lOiBNZWRpYXdpa2kgKEFQQikKICBjb25zb2xlLm9wZW5zaGlmdC5pby9pY29uQ2xhc3M6IGljb24tbWVkaWF3aWtpCiAgcHJvdmlkZXJEaXNwbGF5TmFtZTogIlJlZCBIYXQsIEluYy4iCnBsYW5zOgogIC0gbmFtZTogZGVmYXVsdAogICAgZGVzY3JpcHRpb246IEFuIEFQQiB0aGF0IGRlcGxveXMgTWVkaWFXaWtpCiAgICBmcmVlOiBUcnVlCiAgICBtZXRhZGF0YToKICAgICAgZGlzcGxheU5hbWU6IERlZmF1bHQKICAgICAgbG9uZ0Rlc2NyaXB0aW9uOiBUaGlzIHBsYW4gZGVwbG95cyBhIHNpbmdsZSBtZWRpYXdpa2kgaW5zdGFuY2Ugd2l0aG91dCBhIERCCiAgICAgIGNvc3Q6ICQwLjAwCiAgICBwYXJhbWV0ZXJzOgogICAgICAtIG5hbWU6IG1lZGlhd2lraV9kYl9zY2hlbWEKICAgICAgICBkZWZhdWx0OiBtZWRpYXdpa2kKICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIERCIFNjaGVtYQogICAgICAgIHJlcXVpcmVkOiBUcnVlCiAgICAgIC0gbmFtZTogbWVkaWF3aWtpX3NpdGVfbmFtZQogICAgICAgIGRlZmF1bHQ6IE1lZGlhV2lraQogICAgICAgIHR5cGU6IHN0cmluZwogICAgICAgIHRpdGxlOiBNZWRpYXdpa2kgU2l0ZSBOYW1lCiAgICAgICAgcmVxdWlyZWQ6IFRydWUKICAgICAgICB1cGRhdGFibGU6IFRydWUKICAgICAgLSBuYW1lOiBtZWRpYXdpa2lfc2l0ZV9sYW5nCiAgICAgICAgZGVmYXVsdDogZW4KICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIFNpdGUgTGFuZ3VhZ2UKICAgICAgICByZXF1aXJlZDogVHJ1ZQogICAgICAtIG5hbWU6IG1lZGlhd2lraV9hZG1pbl91c2VyCiAgICAgICAgZGVmYXVsdDogYWRtaW4KICAgICAgICB0eXBlOiBzdHJpbmcKICAgICAgICB0aXRsZTogTWVkaWF3aWtpIEFkbWluIFVzZXIKICAgICAgICByZXF1aXJlZDogVHJ1ZQogICAgICAtIG5hbWU6IG1lZGlhd2lraV9hZG1pbl9wYXNzCiAgICAgICAgdHlwZTogc3RyaW5nCiAgICAgICAgdGl0bGU6IE1lZGlhd2lraSBBZG1pbiBVc2VyIFBhc3N3b3JkCiAgICAgICAgcmVxdWlyZWQ6IFRydWUKICAgICAgICBkaXNwbGF5X3R5cGU6IHBhc3N3b3JkCg=="
	type history struct {
		History []map[string]string `json:"history"`
	}
	cases := []struct {
		Name     string
		Response history
		Validate func(t *testing.T, spec *bundle.Spec)
	}{
		{
			Name: "test spec parsed correctly and runtime version 1 and spec version 1.0 when no Label present",
			Response: history{History: []map[string]string{
				{
					"v1Compatibility": fmt.Sprintf("{\"config\":{\"Labels\":{\"build-date\":\"20170801\",\"com.redhat.apb.spec\":\"%s\",\"com.redhat.apb.version\":\"0.1.0\"}}}", testApbSpec),
				},
			}},
			Validate: func(t *testing.T, spec *bundle.Spec) {
				if spec.Runtime != 1 {
					t.Fatalf("Expected the runtime to be %v but it was %v", 1, spec.Runtime)
				}
				if spec.Version != "1.0" {
					t.Fatalf("Expected the version to be %v but it was %v", "1.0", spec.Version)
				}
			},
		},
		{
			Name: "test spec parsed correctly and runtime version 2 and spec version 1.0 when apb Label present",
			Response: history{History: []map[string]string{
				{
					"v1Compatibility": fmt.Sprintf("{\"config\":{\"Labels\":{\"build-date\":\"20170801\",\"com.redhat.apb.spec\":\"%s\",\"com.redhat.apb.version\":\"0.1.0\",\"com.redhat.apb.runtime\":\"2\"}}}", testApbSpec),
				},
			}},
			Validate: func(t *testing.T, spec *bundle.Spec) {
				if spec.Runtime != 2 {
					t.Fatalf("Expected the runtime to be %v but it was %v", 2, spec.Runtime)
				}
				if spec.Version != "1.0" {
					t.Fatalf("Expected the version to be %v but it was %v", "1.0", spec.Version)
				}
			},
		},
		{
			Name: "test spec parsed correctly and runtime version 2 and spec version 1.0 when bundle Label present",
			Response: history{History: []map[string]string{
				{
					"v1Compatibility": fmt.Sprintf("{\"config\":{\"Labels\":{\"build-date\":\"20170801\",\"com.redhat.apb.spec\":\"%s\",\"com.redhat.apb.version\":\"0.1.0\",\"com.redhat.bundle.runtime\":\"2\"}}}", testApbSpec),
				},
			}},
			Validate: func(t *testing.T, spec *bundle.Spec) {
				if spec.Runtime != 2 {
					t.Fatalf("Expected the runtime to be %v but it was %v", 2, spec.Runtime)
				}
				if spec.Version != "1.0" {
					t.Fatalf("Expected the version to be %v but it was %v", "1.0", spec.Version)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			b, err := json.Marshal(tc.Response)
			if err != nil {
				t.Fatalf("failed to marshal response from test case %v", err)
			}
			spec, err := imageToSpec(b, "maleck13/3scale-apb")
			if err != nil {
				t.Fatal(err)
			}
			if tc.Validate != nil {
				tc.Validate(t, spec)
			}
		})
	}
}
