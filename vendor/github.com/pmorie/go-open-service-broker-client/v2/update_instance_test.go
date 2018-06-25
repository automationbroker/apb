package v2

import (
	"fmt"
	"net/http"
	"testing"
)

func defaultUpdateInstanceRequest() *UpdateInstanceRequest {
	return &UpdateInstanceRequest{
		InstanceID: testInstanceID,
		ServiceID:  testServiceID,
		PlanID:     strPtr(testPlanID),
	}
}

func defaultAsyncUpdateInstanceRequest() *UpdateInstanceRequest {
	r := defaultUpdateInstanceRequest()
	r.AcceptsIncomplete = true
	return r
}

const successUpdateInstanceRequestBody = `{"service_id":"test-service-id","plan_id":"test-plan-id"}`

const successUpdateInstanceResponseBody = `{}`
const successUpdateInstanceResponseBodyWithNewDashboardURL = `{"dashboard_url":"http://updated.com"}`

func successUpdateInstanceResponse() *UpdateInstanceResponse {
	return &UpdateInstanceResponse{}
}

const successAsyncUpdateInstanceResponseBody = `{
  "operation": "test-operation-key"
}`
const successAsyncUpdateInstanceResponseBodyWithNewDashboardURL = `{
	"dashboard_url": "http://updated.com",
	"operation": "test-operation-key"
}`

func successUpdateInstanceResponseWithDashboard() *UpdateInstanceResponse {
	r := successUpdateInstanceResponse()
	url := "http://updated.com"
	r.DashboardURL = &url
	return r
}

func successUpdateInstanceResponseAsync() *UpdateInstanceResponse {
	r := successUpdateInstanceResponse()
	r.Async = true
	r.OperationKey = &testOperation
	return r
}

func successUpdateInstanceResponeAsyncWithDashboard() *UpdateInstanceResponse {
	r := successUpdateInstanceResponseAsync()
	url := "http://updated.com"
	r.DashboardURL = strPtr(url)
	return r
}

const contextUpdateInstanceRequestBody = `{"service_id":"test-service-id","plan_id":"test-plan-id","context":{"foo":"bar"}}`

const previousValuesUpdateInstanceRequestBody = `{"service_id":"test-service-id","plan_id":"test-plan-id","previous_values":{"plan_id":"previous-plan-id"}}`

func TestUpdateInstanceInstance(t *testing.T) {
	cases := []struct {
		name                string
		version             APIVersion
		enableAlpha         bool
		originatingIdentity *OriginatingIdentity
		request             *UpdateInstanceRequest
		httpChecks          httpChecks
		httpReaction        httpReaction
		expectedResponse    *UpdateInstanceResponse
		expectedErrMessage  string
		expectedErr         error
	}{
		{
			name: "invalid request",
			request: func() *UpdateInstanceRequest {
				r := defaultUpdateInstanceRequest()
				r.InstanceID = ""
				return r
			}(),
			expectedErrMessage: "instanceID is required",
		},
		{
			name: "success - ok",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name:    "success - async",
			request: defaultAsyncUpdateInstanceRequest(),
			httpChecks: httpChecks{
				params: map[string]string{
					AcceptsIncomplete: "true",
				},
			},
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   successAsyncUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponseAsync(),
		},
		{
			name:    "accepted with malformed response",
			request: defaultAsyncUpdateInstanceRequest(),
			httpChecks: httpChecks{
				params: map[string]string{
					AcceptsIncomplete: "true",
				},
			},
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   malformedResponse,
			},
			expectedErrMessage: "Status: 202; ErrorMessage: <nil>; Description: <nil>; ResponseError: unexpected end of JSON input",
		},
		{
			name: "http error",
			httpReaction: httpReaction{
				err: fmt.Errorf("http error"),
			},
			expectedErrMessage: "http error",
		},
		{
			name: "202 with no async support",
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   successAsyncUpdateInstanceResponseBody,
			},
			expectedErrMessage: "Status: 202; ErrorMessage: <nil>; Description: <nil>; ResponseError: <nil>",
		},
		{
			name: "200 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   malformedResponse,
			},
			expectedErrMessage: "Status: 200; ErrorMessage: <nil>; Description: <nil>; ResponseError: unexpected end of JSON input",
		},
		{
			name: "500 with malformed response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   malformedResponse,
			},
			expectedErrMessage: "Status: 500; ErrorMessage: <nil>; Description: <nil>; ResponseError: unexpected end of JSON input",
		},
		{
			name: "500 with conventional failure response",
			httpReaction: httpReaction{
				status: http.StatusInternalServerError,
				body:   conventionalFailureResponseBody,
			},
			expectedErr: testHTTPStatusCodeError(),
		},
		{
			name:    "context - 2.12",
			version: Version2_12(),
			request: func() *UpdateInstanceRequest {
				r := defaultUpdateInstanceRequest()
				r.Context = map[string]interface{}{
					"foo": "bar",
				}
				return r
			}(),
			httpChecks: httpChecks{
				body: contextUpdateInstanceRequestBody,
			},
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name: "context - 2.11",
			request: func() *UpdateInstanceRequest {
				r := defaultUpdateInstanceRequest()
				r.Context = map[string]interface{}{
					"foo": "bar",
				}
				return r
			}(),
			httpChecks: httpChecks{
				body: successUpdateInstanceRequestBody,
			},
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name: "previous values",
			request: func() *UpdateInstanceRequest {
				r := defaultUpdateInstanceRequest()
				r.PreviousValues = &PreviousValues{
					PlanID: "previous-plan-id",
				}
				return r
			}(),
			httpChecks: httpChecks{
				body: previousValuesUpdateInstanceRequestBody,
			},
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name:                "originating identity included",
			version:             Version2_13(),
			originatingIdentity: testOriginatingIdentity,
			httpChecks:          httpChecks{headers: map[string]string{OriginatingIdentityHeader: testOriginatingIdentityHeaderValue}},
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name:                "originating identity excluded",
			version:             Version2_13(),
			originatingIdentity: nil,
			httpChecks:          httpChecks{headers: map[string]string{OriginatingIdentityHeader: ""}},
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name:                "originating identity not sent unless API version >= 2.13",
			version:             Version2_12(),
			originatingIdentity: testOriginatingIdentity,
			httpChecks:          httpChecks{headers: map[string]string{OriginatingIdentityHeader: ""}},
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBody,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name:        "success with updated dashboard url - ok if alpha API features are enabled",
			version:     LatestAPIVersion(),
			enableAlpha: true,
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBodyWithNewDashboardURL,
			},
			expectedResponse: successUpdateInstanceResponseWithDashboard(),
		},
		{
			name:        "success with updated dashboard url - async if alpha API features are enabled",
			version:     LatestAPIVersion(),
			enableAlpha: true,
			request:     defaultAsyncUpdateInstanceRequest(),
			httpChecks: httpChecks{
				params: map[string]string{
					AcceptsIncomplete: "true",
				},
			},
			httpReaction: httpReaction{
				status: http.StatusAccepted,
				body:   successAsyncUpdateInstanceResponseBodyWithNewDashboardURL,
			},
			expectedResponse: successUpdateInstanceResponeAsyncWithDashboard(),
		},
		{
			name:        "dashboard url not sent unless alpha API features enabled",
			version:     LatestAPIVersion(),
			enableAlpha: false,
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBodyWithNewDashboardURL,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
		{
			name:        "dashboard url not sent unless latest version of the API is used",
			version:     Version2_12(),
			enableAlpha: true,
			httpReaction: httpReaction{
				status: http.StatusOK,
				body:   successUpdateInstanceResponseBodyWithNewDashboardURL,
			},
			expectedResponse: successUpdateInstanceResponse(),
		},
	}

	for _, tc := range cases {
		if tc.request == nil {
			tc.request = defaultUpdateInstanceRequest()
		}

		tc.request.OriginatingIdentity = tc.originatingIdentity

		if tc.httpChecks.URL == "" {
			tc.httpChecks.URL = "/v2/service_instances/test-instance-id"
		}

		if tc.httpChecks.body == "" {
			tc.httpChecks.body = "{\"service_id\":\"test-service-id\",\"plan_id\":\"test-plan-id\"}"
		}

		if tc.version.label == "" {
			tc.version = Version2_11()
		}

		klient := newTestClient(t, tc.name, tc.version, tc.enableAlpha, tc.httpChecks, tc.httpReaction)

		response, err := klient.UpdateInstance(tc.request)

		doResponseChecks(t, tc.name, response, err, tc.expectedResponse, tc.expectedErrMessage, tc.expectedErr)
	}
}

func TestValidateUpdateInstanceRequest(t *testing.T) {
	cases := []struct {
		name    string
		request *UpdateInstanceRequest
		valid   bool
	}{
		{
			name:    "valid",
			request: defaultUpdateInstanceRequest(),
			valid:   true,
		},
		{
			name: "missing instance ID",
			request: func() *UpdateInstanceRequest {
				r := defaultUpdateInstanceRequest()
				r.InstanceID = ""
				return r
			}(),
			valid: false,
		},
		{
			name: "missing service ID",
			request: func() *UpdateInstanceRequest {
				r := defaultUpdateInstanceRequest()
				r.InstanceID = "instanceID"
				r.ServiceID = ""
				return r
			}(),
			valid: false,
		},
	}

	for _, tc := range cases {
		err := validateUpdateInstanceRequest(tc.request)
		if err != nil {
			if tc.valid {
				t.Errorf("%v: expected valid, got error: %v", tc.name, err)
			}
		} else if !tc.valid {
			t.Errorf("%v: expected invalid, got valid", tc.name)
		}
	}
}
