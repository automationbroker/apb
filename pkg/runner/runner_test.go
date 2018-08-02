package runner

import (
	"github.com/automationbroker/bundle-lib/bundle"
	"testing"
)

func TestContains(t *testing.T) {
	// test case table
	testCases := []struct {
		name     string
		list     []string
		target   string
		contains bool
	}{
		{
			name:     "test contains with target",
			list:     []string{"foo", "bar"},
			target:   "foo",
			contains: true,
		},
		{
			name:     "test contains without target",
			list:     []string{"foo", "bar"},
			target:   "leto",
			contains: false,
		},
		{
			name:     "test contains with empty list",
			list:     []string{},
			target:   "foo",
			contains: false,
		},
		{
			name:     "test contains with empty target",
			list:     []string{"foo", "bar"},
			target:   "",
			contains: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			/* Testing logic */
			if contains(tc.list, tc.target) && !tc.contains {
				t.Fatalf("expected contains to fail but function succeeded")
				return
			}
			if !contains(tc.list, tc.target) && tc.contains {
				t.Fatalf("expected contains to succeed but function failed")
				return
			}
		})
	}
}

func TestPruneInput(t *testing.T) {
	// test case table
	testCases := []struct {
		name      string
		param     bundle.ParameterDescriptor
		input     string
		shouldErr bool
	}{
		{
			name: "test valid string",
			param: bundle.ParameterDescriptor{
				Type: "string",
			},
			input:     "valid-string",
			shouldErr: false,
		},
		{
			name: "test valid enum",
			param: bundle.ParameterDescriptor{
				Type: "enum",
			},
			input:     "valid-enum",
			shouldErr: false,
		},
		{
			name: "test valid boolean",
			param: bundle.ParameterDescriptor{
				Type: "boolean",
			},
			input:     "true",
			shouldErr: false,
		},
		{
			name: "test invalid boolean",
			param: bundle.ParameterDescriptor{
				Type: "boolean",
			},
			input:     "foo",
			shouldErr: true,
		},
		{
			name: "test valid integer",
			param: bundle.ParameterDescriptor{
				Type: "integer",
			},
			input:     "20",
			shouldErr: false,
		},
		{
			name: "test invalid integer",
			param: bundle.ParameterDescriptor{
				Type: "integer",
			},
			input:     "foo",
			shouldErr: true,
		},
		{
			name: "test invalid integer",
			param: bundle.ParameterDescriptor{
				Type: "integer",
			},
			input:     "20foo",
			shouldErr: true,
		},
		{
			name: "test invalid number",
			param: bundle.ParameterDescriptor{
				Type: "number",
			},
			input:     "foo",
			shouldErr: true,
		},
		{
			name: "test valid number",
			param: bundle.ParameterDescriptor{
				Type: "number",
			},
			input:     "22.4",
			shouldErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			/* Testing logic */
			output, err := pruneInput(tc.input, tc.param)
			if err != nil && !tc.shouldErr {
				t.Fatalf("got unexpected error [%v]", err)
				return
			}
			if tc.shouldErr && output != nil {
				t.Fatalf("Expected error but got output [%v]", output)
				return
			}
			switch output.(type) {
			case int:
				if tc.param.Type != "int" && tc.param.Type != "integer" {
					t.Fatalf("got unexpected output type [%v]. expected [int]", tc.param.Type)
					return
				}
			case string:
				if tc.param.Type != "string" && tc.param.Type != "enum" {
					t.Fatalf("got unexpected output type [%v]. expected [string]", tc.param.Type)
					return
				}
			case float64:
				if tc.param.Type != "float" && tc.param.Type != "number" {
					t.Fatalf("got unexpected output type [%v]. expected [float64]", tc.param.Type)
					return
				}
			case bool:
				if tc.param.Type != "bool" && tc.param.Type != "boolean" {
					t.Fatalf("got unexpected output type [%v]. expected [bool]", tc.param.Type)
					return
				}
			}
		})
	}
}

func TestCreateExtraVars(t *testing.T) {
	// test case table
	testCases := []struct {
		name      string
		targetNs  string
		params    bundle.Parameters
		shouldErr bool
	}{
		{
			name:     "test valid params",
			targetNs: "foo",
			params: &bundle.Parameters{
				"test-param": true,
			},
			plan:      bundle.Plan{},
			shouldErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			/* Testing logic */
			vars, err := createExtraVars(tc.targetNs, tc.params, tc.plan)
		})
	}
}
