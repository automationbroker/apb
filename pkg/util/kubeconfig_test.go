package util

import (
	"testing"
)

func TestGetCurrentNamespace(t *testing.T) {
	/*pre test login */
	// test case table
	testCases := []struct {
		name       string
		configPath string
		shouldErr  bool
	}{
		{
			name:       "good kubeconfig",
			configPath: "testdata/config",
			shouldErr:  false,
		},
		{
			name:       "bad kubeconfig path",
			configPath: "testadata/path/that/doesnt/exist",
			shouldErr:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			/* Testing logic */
			namespace := GetCurrentNamespace(tc.configPath)
			if namespace == "" && !tc.shouldErr {
				t.Fatalf("Didn't find namespace and error not expected")
				return
			}
			if namespace != "foo-ns" && !tc.shouldErr {
				t.Fatalf("Failed to find expected namespace foo-ns. Got [%v]", namespace)
				return
			}
			return
		})
	}
}
