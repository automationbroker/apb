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

func createHook(p, n string, targets []string, role string) error {
	return nil
}

func destroyHook(p, n string, targets []string) error {
	return nil
}

func TestAddPreCreateSandbox(t *testing.T) {
	p := provider{}
	p.addPreCreateSandbox(createHook)
	if len(p.preSandboxCreate) == 0 {
		t.Fatal("sandbox hooks was not added")
	}
}

func TestAddPostCreateSandbox(t *testing.T) {
	p := provider{}
	p.addPostCreateSandbox(createHook)
	if len(p.postSandboxCreate) == 0 {
		t.Fatal("sandbox hooks was not added")
	}

}

func TestAddPreDestroySandbox(t *testing.T) {
	p := provider{}
	p.addPreDestroySandbox(destroyHook)
	if len(p.preSandboxDestroy) == 0 {
		t.Fatal("sandbox hooks was not added")
	}
}

func TestAddPostDestroySandbox(t *testing.T) {
	p := provider{}
	p.addPostDestroySandbox(destroyHook)
	if len(p.postSandboxDestroy) == 0 {
		t.Fatal("sandbox hooks was not added")
	}
}
