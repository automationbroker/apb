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
	"encoding/base64"
	"fmt"
	"testing"

	schema "github.com/lestrrat/go-jsschema"
	ft "github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
	"strings"
)

var emailPlanParams = []ParameterDescriptor{
	{
		Name:        "email_address",
		Title:       "Email Address",
		Type:        "enum",
		Description: "example enum parameter",
		Enum:        []string{"google@gmail.com", "redhat@redhat.com"},
		Default:     "google@gmail.com",
		Updatable:   true,
	},
	{
		Name:        "password",
		Title:       "Password",
		Type:        "string",
		Description: "example string parameter with a display type",
		DisplayType: "password",
	},
	{
		Name:         "first_name",
		Title:        "First Name",
		Type:         "string",
		Description:  "example grouped string parameter",
		DisplayGroup: "User Information",
	},
	{
		Name:         "last_name",
		Title:        "Last Name",
		Type:         "string",
		Description:  "example grouped string parameter",
		DisplayGroup: "User Information",
	},
}

var emailPlanBindParams = []ParameterDescriptor{
	{
		Name:        "bind_param_1",
		Title:       "Bind Param 1",
		Type:        "string",
		Description: "Bind Param 1",
		DisplayType: "text",
	},
	{
		Name:         "bind_param_2",
		Title:        "Bind Param 2",
		Type:         "string",
		Description:  "Bind Param 2",
		DisplayGroup: "Bind Group 1",
	},
	{
		Name:         "bind_param_3",
		Title:        "Bind Param 3",
		Type:         "string",
		Description:  "Bind Param 3",
		DisplayGroup: "Bind Group 1",
	},
}

var pl = Plan{
	ID:             "55822a921d2c4858fe6e58f5522429c2", // md5(dh-sns-apb-dev)
	Name:           PlanName,
	Description:    PlanDescription,
	Metadata:       PlanMetadata,
	Free:           PlanFree,
	Bindable:       PlanBindable,
	Parameters:     emailPlanParams,
	BindParameters: emailPlanBindParams,
}

func TestEnumIsCopied(t *testing.T) {

	schemaObj, _ := parametersToSchema(pl)

	emailParam := schemaObj.ServiceInstance.Create["parameters"].Properties["email_address"]
	ft.Equal(t, len(emailParam.Enum), 2, "enum mismatch")
	ft.Equal(t, emailParam.Enum[0], "google@gmail.com")
	ft.Equal(t, emailParam.Enum[1], "redhat@redhat.com")

}

func TestEnumIsCopiedForUpdate(t *testing.T) {

	schemaObj, _ := parametersToSchema(pl)

	emailParam := schemaObj.ServiceInstance.Update["parameters"].Properties["email_address"]
	ft.Equal(t, len(emailParam.Enum), 2, "enum mismatch")
	ft.Equal(t, emailParam.Enum[0], "google@gmail.com")
	ft.Equal(t, emailParam.Enum[1], "redhat@redhat.com")

}

func TestConvertPlansToSchema(t *testing.T) {
	schemaPlan, _ := ConvertPlansToSchema([]Plan{pl})
	ft.NotNil(t, schemaPlan, "schema plan is empty")
	emailParam := schemaPlan[0].Schemas.ServiceInstance.Create["parameters"].Properties["email_address"]
	ft.Equal(t, len(emailParam.Enum), 2, "enum mismatch")
	ft.Equal(t, emailParam.Enum[0], "google@gmail.com")
}

func TestUpdateMetadata(t *testing.T) {
	planMetadata := extractBrokerPlanMetadata(pl)
	ft.NotNil(t, planMetadata, "plan metadata is empty")

	verifyInstanceFormDefinition(t, planMetadata, []string{"schemas", "service_instance", "create"})

	updateFormDefnMap := verifyMapPath(t, planMetadata, []string{"schemas", "service_instance", "update"})
	ft.Equal(t, len(updateFormDefnMap), 0, "schemas.service_instance.update is not empty")

	verifyBindingFormDefinition(t, planMetadata, []string{"schemas", "service_binding", "create"})
}

func verifyInstanceFormDefinition(t *testing.T, planMetadata map[string]interface{}, path []string) {

	formDefnMap := verifyMapPath(t, planMetadata, path)
	formDefnMetadata, correctType := formDefnMap["openshift_form_definition"].([]interface{})
	ft.True(t, correctType, strings.Join(path, ".")+" Form definition is of the wrong type")
	ft.NotNil(t, formDefnMetadata, "Form definition is nil")
	ft.Equal(t, len(formDefnMetadata), 3, "Incorrect number of parameters in form definition")

	passwordParam, correctType := formDefnMetadata[1].(formItem)
	ft.True(t, correctType, strings.Join(path, ".")+" Form definition password param is of the wrong type")
	ft.NotNil(t, passwordParam)
	ft.Equal(t, passwordParam.Key, pl.Parameters[1].Name, "Password parameter has the wrong name")
	ft.Equal(t, passwordParam.Type, pl.Parameters[1].DisplayType, "Password parameter display type is incorrect")

	group, correctType := formDefnMetadata[2].(formItem)
	ft.True(t, correctType, strings.Join(path, ".")+" Form definition parameter group is of the wrong type")
	ft.NotNil(t, group, "Parameter group is empty")
	ft.Equal(t, group.Type, "fieldset", "Group form item type is incorrect")
	ft.Equal(t, group.Title, "User Information", "Group form item title is incorrect.")

	groupedItems := group.Items
	ft.NotNil(t, groupedItems, "Group missing parameter items")
	ft.Equal(t, len(groupedItems), 2, "Incorrect number of parameters in group")

	firstNameParam, correctType := groupedItems[0].(string)
	ft.True(t, correctType, "first_name is of the wrong type")
	ft.Equal(t, firstNameParam, pl.Parameters[2].Name, "Incorrect name for first_name")

	lastNameParam, correctType := groupedItems[1].(string)
	ft.True(t, correctType, "last_name is of the wrong type")
	ft.Equal(t, lastNameParam, pl.Parameters[3].Name, "Incorrect name for last_name")
}

func verifyBindingFormDefinition(t *testing.T, planMetadata map[string]interface{}, path []string) {

	formDefnMap := verifyMapPath(t, planMetadata, path)
	formDefnMetadata, correctType := formDefnMap["openshift_form_definition"].([]interface{})
	ft.True(t, correctType, strings.Join(path, ".")+" Form definition is of the wrong type")
	ft.NotNil(t, formDefnMetadata, "Form definition is nil")
	ft.Equal(t, len(formDefnMetadata), 2, "Incorrect number of parameters in form definition")

	bindParam1, correctType := formDefnMetadata[0].(formItem)
	ft.True(t, correctType, strings.Join(path, ".")+" Form definition binding_param_1 is of the wrong type")
	ft.NotNil(t, bindParam1)
	ft.Equal(t, bindParam1.Key, pl.BindParameters[0].Name, "binding_param_1 has the wrong name")
	ft.Equal(t, bindParam1.Type, pl.BindParameters[0].DisplayType, "binding_param_1 display type is incorrect")

	group, correctType := formDefnMetadata[1].(formItem)
	ft.True(t, correctType, strings.Join(path, ".")+" Form definition parameter group is of the wrong type")
	ft.NotNil(t, group, "Parameter group is empty")
	ft.Equal(t, group.Type, "fieldset", "Group form item type is incorrect")
	ft.Equal(t, group.Title, "Bind Group 1", "Group form item title is incorrect.")

	groupedItems := group.Items
	ft.NotNil(t, groupedItems, "Group missing parameter items")
	ft.Equal(t, len(groupedItems), 2, "Incorrect number of parameters in group")

	bindParam2, correctType := groupedItems[0].(string)
	ft.True(t, correctType, "bind_param_2 is of the wrong type")
	ft.Equal(t, bindParam2, pl.BindParameters[1].Name, "Incorrect name for bind_param_2")

	bindParam3, correctType := groupedItems[1].(string)
	ft.True(t, correctType, "bind_param_3 is of the wrong type")
	ft.Equal(t, bindParam3, pl.BindParameters[2].Name, "Incorrect name for bind_param_3")
}

func verifyMapPath(t *testing.T, planMetadata map[string]interface{}, path []string) map[string]interface{} {
	currentMap := planMetadata
	var correctType bool
	for _, jsonKey := range path {
		currentMap, correctType = currentMap[jsonKey].(map[string]interface{})
		ft.True(t, correctType, "incorrectly typed "+jsonKey+" metadata")
		ft.NotNil(t, currentMap, jsonKey+" metadata empty")
	}

	return currentMap
}

func TestParametersToSchema(t *testing.T) {
	decodedyaml, err := base64.StdEncoding.DecodeString(EncodedApb())
	if err != nil {
		t.Fatal(err)
	}

	spec := &Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	schemaObj, _ := parametersToSchema(spec.Plans[0])

	found := false
	for k, p := range schemaObj.ServiceInstance.Create["parameters"].Properties {
		// let's verify the site language
		if k == "mediawiki_site_lang" {
			found = true
			ft.Equal(t, p.Title, "Mediawiki Site Language", "title mismatch")
			ft.True(t, p.Type.Contains(schema.StringType), "type mismatch")
			ft.Equal(t, p.Description, "", "description mismatch")
			ft.Equal(t, p.Default, "en", "default mismatch")
			ft.Equal(t, p.MaxLength.Val, 0, "maxlength mismatch")
			ft.False(t, p.MaxLength.Initialized, "maxlength initialized")
			ft.Equal(t, len(p.Enum), 0, "enum mismatch")
		}
	}
	ft.True(t, found, "no mediawiki_site_lang property found")

	verifyBindParameters(t, schemaObj)
}

func TestUpdateParametersToSchema(t *testing.T) {
	decodedyaml, err := base64.StdEncoding.DecodeString(EncodedApb())
	if err != nil {
		t.Fatal(err)
	}

	spec := &Spec{}
	if err = yaml.Unmarshal(decodedyaml, spec); err != nil {
		t.Fatal(err)
	}
	schemaObj, _ := parametersToSchema(spec.Plans[0])

	found := false
	for k, p := range schemaObj.ServiceInstance.Create["parameters"].Properties {
		// let's verify the site language
		if k == "mediawiki_site_name" {
			found = true
			ft.Equal(t, p.Title, "Mediawiki Site Name", "title mismatch")
			ft.True(t, p.Type.Contains(schema.StringType), "type mismatch")
			ft.Equal(t, p.Description, "", "description mismatch")
			ft.Equal(t, p.Default, "MediaWiki", "default mismatch")
			ft.Equal(t, p.MaxLength.Val, 0, "maxlength mismatch")
			ft.False(t, p.MaxLength.Initialized, "maxlength initialized")
			ft.Equal(t, len(p.Enum), 0, "enum mismatch")
		}
	}
	ft.True(t, found, "no mediawiki_site_lang property found")

	verifyBindParameters(t, schemaObj)
}

func verifyBindParameters(t *testing.T, schemaObj Schema) {
	found1 := false
	found2 := false
	found3 := false
	for k, prop := range schemaObj.ServiceBinding.Create["parameters"].Properties {
		if k == "bind_param_1" {
			found1 = true
			verifyParameter(t, prop, "Bind Param 1", schema.StringType, nil)
		}
		if k == "bind_param_2" {
			found2 = true
			verifyParameter(t, prop, "Bind Param 2", schema.IntegerType, nil)
		}
		if k == "bind_param_3" {
			found3 = true
			verifyParameter(t, prop, "Bind Param 3", schema.StringType, nil)
		}
	}
	ft.True(t, found1, "bind_param_1 not found")
	ft.True(t, found2, "bind_param_2 not found")
	ft.True(t, found3, "bind_param_3 not found")

	found1 = false
	found2 = false
	found3 = false
	for _, k := range schemaObj.ServiceBinding.Create["parameters"].Required {
		if k == "bind_param_1" {
			found1 = true
		}
		if k == "bind_param_2" {
			found2 = true
		}
		if k == "bind_param_3" {
			found3 = true
		}
	}
	ft.True(t, found1, "bind_param_1 not required")
	ft.True(t, found2, "bind_param_2 not required")
	ft.False(t, found3, "bind_param_3 should not be required")
}

func verifyParameter(t *testing.T, property *schema.Schema, paramTitle string, paramType schema.PrimitiveType, paramDefault interface{}) {
	ft.Equal(t, property.Title, paramTitle, "title mismatch"+property.Title+" != "+paramTitle)
	ft.True(t, property.Type.Contains(paramType), paramTitle, "type mismatch")
}

func TestGetType(t *testing.T) {
	// table of testcases
	testCases := []struct {
		jsonType string
		want     schema.PrimitiveType
	}{
		{"string", schema.StringType},
		{"STRING", schema.StringType},
		{"String", schema.StringType},
		{"enum", schema.StringType},
		{"int", schema.IntegerType},
		{"object", schema.ObjectType},
		{"array", schema.ArrayType},
		{"bool", schema.BooleanType},
		{"boolean", schema.BooleanType},
		{"number", schema.NumberType},
		{"nil", schema.NullType},
		{"null", schema.NullType},
		{"biteme", schema.UnspecifiedType},
	}

	// test
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in type %s", tc.want, tc.jsonType), func(t *testing.T) {
			ty, err := getType(tc.jsonType)
			if tc.jsonType == "biteme" && err == nil {
				t.Fatalf("unknown schema types should return an error")
			} else if tc.jsonType == "biteme" && err != nil {
				return
			}
			ft.True(t, ty.Contains(tc.want), "test failed")
		})
	}
}
