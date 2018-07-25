package config

import (
	"github.com/automationbroker/bundle-lib/registries"
	"os"
	"testing"
)

const defaultConfigDir = "testdata/.apb"
const defaultConfigName = "registries"

func TestInitJSONConfigWithData(t *testing.T) {
	/*pre test login */

	// test case table
	testCases := []struct {
		name       string
		configDir  string
		configName string
		regName    string
		regCount   int
		created    bool
	}{
		{
			name:       "test loading data",
			configDir:  defaultConfigDir,
			configName: defaultConfigName,
			regName:    "dylan",
			regCount:   1,
			created:    false,
		},
		{
			name:       "test initializing data",
			configDir:  "testdata/.foo",
			configName: "bar",
			regName:    "",
			regCount:   0,
			created:    true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			/* Testing logic */
			viperConfig, created := InitJSONConfig(tc.configDir, tc.configName)
			var regList []Registry
			if created != tc.created {
				t.Fatalf("expected config creation [%v], got [%v]", tc.created, created)
				return
			}
			viperConfig.UnmarshalKey("Registries", &regList)
			if len(regList) != tc.regCount {
				t.Fatalf("found unexpected number of registries in config")
				return
			}
			if len(regList) > 0 {
				if regList[0].Config.Name != tc.regName {
					t.Fatalf("wrong registry name in config. Expected [%v], found [%v]", regList[0].Config.Name, tc.regName)
					return
				}
			}
			if tc.created == true {
				// Clean up test config
				os.RemoveAll(tc.configDir)
			}
		})
	}
}

func TestUpdateCachedDefaults(t *testing.T) {
	testCases := []struct {
		name                string
		newDefaultNamespace string
	}{
		{
			name:                "test changing default namespace",
			newDefaultNamespace: "new-default-namespace",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defaultSettings := DefaultSettings{}
			viperConfig, created := InitJSONConfig(defaultConfigDir, defaultConfigName)
			if created == true {
				t.Fatalf("unexpected creation value initializing config")
				return
			}
			LoadDefaultSettings(viperConfig, &defaultSettings)
			defaultSettings.BrokerNamespace = tc.newDefaultNamespace
			err := UpdateCachedDefaults(viperConfig, &defaultSettings)
			if err != nil {
				t.Fatalf("unexpected error updating defaults cache")
				return
			}
			LoadDefaultSettings(viperConfig, &defaultSettings)
			if defaultSettings.BrokerNamespace != tc.newDefaultNamespace {
				t.Fatalf("default namespace not updated. Expected [%v], got [%v]", tc.newDefaultNamespace, defaultSettings.BrokerNamespace)
				return
			}
			// Reset defaults
			UpdateCachedDefaults(viperConfig, InitialDefaultSettings())
			LoadDefaultSettings(viperConfig, &defaultSettings)
			if defaultSettings.BrokerNamespace != "openshift-automation-service-broker" {
				t.Fatalf("default namespace not reset. Expected [openshift-automation-service-broker], got [%v]", defaultSettings.BrokerNamespace)
				return
			}
		})
	}
}

func TestUpdateCachedRegistries(t *testing.T) {
	testCases := []struct {
		name       string
		newRegName string
	}{
		{
			name:       "test adding registry",
			newRegName: "atreides",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var regList []Registry
			viperConfig, created := InitJSONConfig(defaultConfigDir, defaultConfigName)
			if created == true {
				t.Fatalf("unexpected creation value initializing config")
				return
			}
			viperConfig.UnmarshalKey("Registries", &regList)
			if len(regList) != 1 {
				t.Fatalf("unexpected count of registries in config [%v], expected [1]", len(regList))
				return
			}
			newReg := Registry{
				Config: registries.Config{
					Name: tc.newRegName,
				},
			}
			regList = append(regList, newReg)
			err := UpdateCachedRegistries(viperConfig, regList)
			if err != nil {
				t.Fatalf("unexpected error updating registry cache")
				return
			}
			if len(regList) != 2 {
				t.Fatalf("registry list was not updated")
				return
			}
			regList = regList[:len(regList)-1]
			err = UpdateCachedRegistries(viperConfig, regList)
			if err != nil {
				t.Fatalf("unexpected error deleting added registry")
				return
			}
		})
	}

}

func TestInitializeDefaultSettings(t *testing.T) {
	// test case table
	testCases := []struct {
		name      string
		namespace string
	}{
		{
			name:      "test default namespace",
			namespace: "openshift-automation-service-broker",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			/* Testing logic */
			settings := InitialDefaultSettings()
			if settings.BrokerNamespace != tc.namespace {
				t.Fatalf("Expected default namespace [%v], got [%v]", tc.namespace, settings.BrokerNamespace)
			}
		})
	}
}
