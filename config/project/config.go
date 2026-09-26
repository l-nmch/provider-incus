// Package project configures the Incus project resource.
package project

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/l-nmch/provider-incus/config/common"
)

// Configure configures the project group.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("incus_project", func(r *config.Resource) {
		r.ShortGroup = "project"
		r.Kind = "Project"
		common.Configure(r)
	})
}
