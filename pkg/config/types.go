package config

import (
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
)

// ProvisionedInstance stores a list of provisioned uuids associated with a bundle
type ProvisionedInstance struct {
	BundleName  string
	InstanceIDs []string
}

// Registry stores a single registry config and references all associated bundle specs
type Registry struct {
	Config registries.Config
	Specs  []*bundle.Spec
}

// DefaultSettings stores default settings for APB tool operation
type DefaultSettings struct {
	BrokerNamespace          string
	BrokerResourceURL        string
	BrokerRouteName          string
	ClusterServiceBrokerName string
	BrokerRouteSuffix        string
}
