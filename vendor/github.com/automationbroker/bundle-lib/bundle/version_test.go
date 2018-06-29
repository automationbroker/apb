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

package bundle

import (
	//"fmt"
	"testing"

	ft "github.com/stretchr/testify/assert"
)

func TestValidateVersion(t *testing.T) {

	// Test Valid Spec Version + Runtime Version
	var testSpec = Spec{
		Version: "1.0.0",
		Runtime: 1,
	}
	ft.True(t, testSpec.ValidateVersion())

	testSpec.Runtime = 2
	ft.True(t, testSpec.ValidateVersion())

	testSpec.Version = "1.0" // Deprecated Spec Version
	ft.True(t, testSpec.ValidateVersion())

	// Test Invalid Spec Versions
	testSpec.Version = "0.9.0" // less than min
	ft.False(t, testSpec.ValidateVersion())

	testSpec.Version = "1.0.1" // greater than max
	ft.False(t, testSpec.ValidateVersion())

	testSpec.Version = "1.0.0" // back to valid version
	ft.True(t, testSpec.ValidateVersion())

	// Test Invalid Runtime Versions
	testSpec.Runtime = 0
	ft.False(t, testSpec.ValidateVersion()) // less than min

	testSpec.Runtime = 3
	ft.False(t, testSpec.ValidateVersion()) // greater than max
}
