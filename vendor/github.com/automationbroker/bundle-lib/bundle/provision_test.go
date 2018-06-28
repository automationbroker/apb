package bundle

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/automationbroker/bundle-lib/runtime"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
)

func TestProvision(t *testing.T) {
	u := uuid.NewUUID()
	testCases := []*struct {
		name            string
		config          ExecutorConfig
		rt              runtime.MockRuntime
		si              ServiceInstance
		extractedCreds  *ExtractedCredentials
		dashboardURL    string
		addExpectations func(rt *runtime.MockRuntime, e Executor)
		validateMessage func([]StatusMessage) bool
	}{
		{
			name:   "provison successfully",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:      "new-spec-id",
					Image:   "new-image",
					FQName:  "new-fq-name",
					Runtime: 2,
				},
				Context: &Context{
					Namespace: "target",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				rt.On("CreateSandbox", mock.Anything, mock.Anything, []string{"target"}, mock.Anything, mock.Anything).Return("service-account-1", "location", nil)
				rt.On("GetRuntime").Return("kubernetes")
				rt.On("CopySecretsToNamespace", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("MasterName", u.String()).Return("new-master-name")
				rt.On("MasterNamespace").Return("new-masternamespace")
				rt.On("StateIsPresent", "new-master-name").Return(false, nil)
				rt.On("RunBundle", mock.Anything).Return(runtime.ExecutionContext{}, nil)
				rt.On("CopyState", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("WatchRunningBundle", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("DestroySandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateSucceeded {
					return false
				}
				return true
			},
		},
		{
			name:   "provison successfully with extracted credentials",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:       "new-spec-id",
					Image:    "new-image",
					FQName:   "new-fq-name",
					Runtime:  2,
					Bindable: true,
				},
				Context: &Context{
					Namespace: "target",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				rt.On("CreateSandbox", mock.Anything, mock.Anything, []string{"target"}, mock.Anything, mock.Anything).Return("service-account-1", "location", nil)
				rt.On("GetRuntime").Return("kubernetes")
				rt.On("CopySecretsToNamespace", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("MasterName", u.String()).Return("new-master-name")
				rt.On("MasterNamespace").Return("new-masternamespace")
				rt.On("StateIsPresent", "new-master-name").Return(false, nil)
				rt.On("RunBundle", mock.Anything).Return(runtime.ExecutionContext{}, nil)
				rt.On("CopyState", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("WatchRunningBundle", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("DestroySandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				rt.On("ExtractCredentials", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"test": "testingcreds"}`), nil)
				rt.On("CreateExtractedCredential", u.String(), mock.Anything, map[string]interface{}{"test": "testingcreds"}, map[string]string{"apbAction": "provision", "apbName": "new-fq-name"}).Return(nil)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateSucceeded {
					return false
				}
				return true
			},
			extractedCreds: &ExtractedCredentials{
				Credentials: map[string]interface{}{"test": "testingcreds"},
			},
		},
		{
			name:   "provison successfully with extracted credentials and dashboard url",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:       "new-spec-id",
					Image:    "new-image",
					FQName:   "new-fq-name",
					Runtime:  2,
					Bindable: true,
				},
				Context: &Context{
					Namespace: "target",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				ex, ok := e.(*executor)
				if !ok {
					t.Fail()
				}
				rt.On("CreateSandbox", mock.Anything, mock.Anything, []string{"target"}, mock.Anything, mock.Anything).Return("service-account-1", "location", nil)
				rt.On("GetRuntime").Return("kubernetes")
				rt.On("CopySecretsToNamespace", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("MasterName", u.String()).Return("new-master-name")
				rt.On("MasterNamespace").Return("new-masternamespace")
				rt.On("StateIsPresent", "new-master-name").Return(false, nil)
				rt.On("RunBundle", mock.Anything).Return(runtime.ExecutionContext{}, nil)
				rt.On("CopyState", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("WatchRunningBundle", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					ex.updateDescription("dashboard url", "https://url.com")
				}).Return(nil)
				rt.On("DestroySandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				rt.On("ExtractCredentials", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"test": "testingcreds"}`), nil)
				rt.On("CreateExtractedCredential", u.String(), mock.Anything, map[string]interface{}{"test": "testingcreds"}, map[string]string{"apbAction": "provision", "apbName": "new-fq-name"}).Return(nil)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 3 {
					return false
				}
				first := m[0]
				second := m[1]
				third := m[2]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateInProgress {
					return false
				}
				if third.State != StateSucceeded {
					return false
				}
				return true
			},
			extractedCreds: &ExtractedCredentials{
				Credentials: map[string]interface{}{"test": "testingcreds"},
			},
		},
		{
			name: "provison successfully skip ns",
			config: ExecutorConfig{
				SkipCreateNS: true,
			},
			rt: *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:      "new-spec-id",
					Image:   "new-image",
					FQName:  "new-fq-name",
					Runtime: 2,
				},
				Context: &Context{
					Namespace: "target",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				rt.On("CreateSandbox", mock.Anything, "target", []string{"target"}, mock.Anything, mock.Anything).Return("service-account-1", "location", nil)
				rt.On("GetRuntime").Return("kubernetes")
				rt.On("CopySecretsToNamespace", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("MasterName", u.String()).Return("new-master-name")
				rt.On("MasterNamespace").Return("new-masternamespace")
				rt.On("StateIsPresent", "new-master-name").Return(false, nil)
				rt.On("RunBundle", mock.Anything).Return(runtime.ExecutionContext{}, nil)
				rt.On("CopyState", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("WatchRunningBundle", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("DestroySandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateSucceeded {
					return false
				}
				return true
			},
		},
		{
			name:   "provison unsuccessfully no image",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:      "new-spec-id",
					Image:   "",
					FQName:  "new-fq-name",
					Runtime: 2,
				},
				Context: &Context{
					Namespace: "target",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				if second.Error.Error() != "No image field found on instance.Spec" {
					return false
				}
				return true
			},
		},
		{
			name: "provison unsuccessfully sandbox fail",
			config: ExecutorConfig{
				SkipCreateNS: true,
			},
			rt: *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:      "new-spec-id",
					Image:   "new-image",
					FQName:  "new-fq-name",
					Runtime: 2,
				},
				Context: &Context{
					Namespace: "target",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				rt.On("CreateSandbox", mock.Anything, "target", []string{"target"}, mock.Anything, mock.Anything).Return("", "", fmt.Errorf("unknown error"))
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
		},
		{
			name: "provison unsuccessfully no location or targets",
			config: ExecutorConfig{
				SkipCreateNS: true,
			},
			rt: *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:      "new-spec-id",
					Image:   "new-image",
					FQName:  "new-fq-name",
					Runtime: 2,
				},
				Context: &Context{
					Namespace: "",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				rt.On("CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("service-account-1", "", nil)
				rt.On("GetRuntime").Return("kubernetes")
				rt.On("DestroySandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
		},
		{
			name:   "provison unsuccessfully to create extracted credentials",
			config: ExecutorConfig{},
			rt:     *new(runtime.MockRuntime),
			si: ServiceInstance{
				ID: u,
				Spec: &Spec{
					ID:       "new-spec-id",
					Image:    "new-image",
					FQName:   "new-fq-name",
					Runtime:  2,
					Bindable: true,
				},
				Context: &Context{
					Namespace: "target",
					Platform:  "kubernetes",
				},
				Parameters: &Parameters{"test-param": true},
			},
			addExpectations: func(rt *runtime.MockRuntime, e Executor) {
				rt.On("CreateSandbox", mock.Anything, mock.Anything, []string{"target"}, mock.Anything, mock.Anything).Return("service-account-1", "location", nil)
				rt.On("GetRuntime").Return("kubernetes")
				rt.On("CopySecretsToNamespace", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("MasterName", u.String()).Return("new-master-name")
				rt.On("MasterNamespace").Return("new-masternamespace")
				rt.On("StateIsPresent", "new-master-name").Return(false, nil)
				rt.On("RunBundle", mock.Anything).Return(runtime.ExecutionContext{}, nil)
				rt.On("CopyState", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("WatchRunningBundle", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				rt.On("DestroySandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				rt.On("ExtractCredentials", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"test": "testingcreds"}`), nil)
				rt.On("CreateExtractedCredential", u.String(), mock.Anything, map[string]interface{}{"test": "testingcreds"}, map[string]string{"apbAction": "provision", "apbName": "new-fq-name"}).Return(fmt.Errorf("unable to create extracted creds"))
			},
			validateMessage: func(m []StatusMessage) bool {
				if len(m) != 2 {
					return false
				}
				first := m[0]
				second := m[1]
				if first.State != StateInProgress {
					return false
				}
				if second.State != StateFailed {
					return false
				}
				return true
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtime.Provider = &tc.rt
			e := NewExecutor(tc.config)
			if tc.addExpectations != nil {
				tc.addExpectations(&tc.rt, e)
			}
			s := e.Provision(&tc.si)
			m := []StatusMessage{}
			for mess := range s {
				m = append(m, mess)
			}
			if !tc.validateMessage(m) {
				t.Fatalf("invalid messages - %#v", m)
			}
			if tc.extractedCreds != nil {
				if !reflect.DeepEqual(e.ExtractedCredentials(), tc.extractedCreds) {
					t.Fatalf("Invalid extracted credentials\nexpected: %#+v\n\nactual: %#+v", tc.extractedCreds, e.ExtractedCredentials())
				}
			}
			if tc.dashboardURL != "" {
				if e.DashboardURL() == tc.dashboardURL {
					t.Fatalf("Invalid dashboard url\nexpected: %v\nactual: %v", tc.dashboardURL, e.DashboardURL())
				}
			}
		})
	}
}
