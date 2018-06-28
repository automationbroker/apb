package clients

import (
	"fmt"
	"reflect"
	"testing"

	authapi "github.com/openshift/api/authorization/v1"
	routeapi "github.com/openshift/api/route/v1"
	authfake "github.com/openshift/client-go/authorization/clientset/versioned/fake"
	routefake "github.com/openshift/client-go/route/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

func TestRoute(t *testing.T) {
	o, err := Openshift()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name      string
		host      string
		route     *routeapi.Route
		namespace string
	}{
		{
			name: "get route",
			host: "foo-route.example.com",
			route: &routeapi.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "route1",
					Namespace: "ns1",
				},
				Spec: routeapi.RouteSpec{
					Host: "foo-route.example.com",
				},
				Status: routeapi.RouteStatus{},
			},
			namespace: "ns1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := routefake.NewSimpleClientset(tc.route)
			o.routeClient = c.Route()
			routeList, err := o.Route().Routes(tc.namespace).List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("error getting route")
				return
			}
			if len(routeList.Items) == 0 {
				t.Fatalf("no routes returned")
				return
			}
			if routeList.Items[0].Spec.Host != tc.host {
				t.Fatalf("route host did not match. Expected: [%v], Got [%v]", tc.host, routeList.Items[0].Spec.Host)
				return
			}
		})
	}
}

func TestOpenshiftSubjectRulesReview(t *testing.T) {
	o, err := Openshift()
	if err != nil {
		t.Fail()
	}

	testCases := []struct {
		name               string
		subjectRulesReview *authapi.SubjectRulesReview
		rules              []authapi.PolicyRule
		user               string
		groups             []string
		scopes             authapi.OptionalScopes
		namespace          string
		shouldErr          bool
	}{
		{
			name: "get rules",
			subjectRulesReview: &authapi.SubjectRulesReview{
				Spec: authapi.SubjectRulesReviewSpec{
					User: "test-users",
				},
				Status: authapi.SubjectRulesReviewStatus{
					Rules: []authapi.PolicyRule{
						authapi.PolicyRule{
							Verbs:         []string{"create"},
							Resources:     []string{"deployments"},
							ResourceNames: []string{"deployment"},
							APIGroups:     []string{"v1"},
						},
					},
				},
			},
			rules: []authapi.PolicyRule{
				authapi.PolicyRule{
					Verbs:         []string{"create"},
					Resources:     []string{"deployments"},
					ResourceNames: []string{"deployment"},
					APIGroups:     []string{"v1"},
				},
			},
			user:      "test-user",
			namespace: "ns1",
		},
		{
			name:      "get rules error",
			user:      "test-user",
			namespace: "ns1",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := authfake.Clientset{}
			c.PrependReactor("create", "subjectrulesreviews", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				if tc.shouldErr {
					return true, nil, fmt.Errorf("should error")
				}
				ca, ok := action.(clientgotesting.CreateActionImpl)
				if !ok {
					return true, nil, fmt.Errorf("can not get create action")
				}
				s, ok := ca.Object.(*authapi.SubjectRulesReview)
				if !ok {
					return true, nil, fmt.Errorf("can not get subject rules review")
				}
				if s.Spec.User != tc.user {
					t.Fatalf("invalid user in spec of request: %v", s.Spec.User)
					return true, nil, fmt.Errorf("invalid spec")
				}
				if ca.Namespace != tc.namespace {
					t.Fatalf("invalid namespace in spec of request: %v", ca.Namespace)
					return true, nil, fmt.Errorf("invalid spec")
				}
				if !reflect.DeepEqual(s.Spec.Groups, tc.groups) {
					t.Fatalf("invalid groups in spec of request: %#v expected: %#v", s.Spec.Groups, tc.groups)
					return true, nil, fmt.Errorf("invalid spec")
				}
				if !reflect.DeepEqual(s.Spec.Scopes, tc.scopes) {
					t.Fatalf("invalid groups in spec of request: %#v expected: %#v", s.Spec.Scopes, tc.scopes)
					return true, nil, fmt.Errorf("invalid spec")
				}
				return true, tc.subjectRulesReview, nil

			})
			o.authClient = c.AuthorizationV1()
			rules, err := o.SubjectRulesReview(tc.user, tc.groups, tc.scopes, tc.namespace)
			if err != nil && !tc.shouldErr {
				t.Fatalf("unknown error - %v", err)
				return
			}
			if err != nil && tc.shouldErr {
				return
			}
			if !reflect.DeepEqual(rules, tc.rules) {
				t.Fatalf("\nActual Rules: %#v\n\nExpected Rules: %#v\n", rules, tc.rules)
				return
			}
		})
	}
}
