package runner

import (
	"github.com/automationbroker/bundle-lib/bundle"
	"github.com/automationbroker/bundle-lib/registries"
)

// Registry stores a single registry config and references all associated bundle specs
type Registry struct {
	Config registries.Config
	Specs  []*bundle.Spec
}
