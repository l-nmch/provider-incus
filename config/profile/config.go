// Package profile configures the Incus profile resource.
package profile

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/l-nmch/provider-incus/config/common"
)

// Configure configures the profile group.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("incus_profile", func(r *config.Resource) {
		r.ShortGroup = "profile"
		r.Kind = "Profile"
		common.Configure(r)
	})
}
