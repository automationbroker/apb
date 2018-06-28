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

import "testing"

func TestKubernetesShouldJoinNetworks(t *testing.T) {
	k := kubernetes{}
	s := k.getRuntime()
	if s != "kubernetes" {
		t.Fatal("runtime does not match kubernetes")
	}
}

func TestKubernetesGetRuntime(t *testing.T) {
	k := kubernetes{}

	jn, postCreateHook, postDestroyHook := k.shouldJoinNetworks()
	if jn || postCreateHook != nil || postDestroyHook != nil {
		t.Fatal("should join networks, or sand box hooks were not nil.")
	}
}
