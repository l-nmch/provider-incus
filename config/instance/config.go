// Package instance configures the Incus instance and instance snapshot
// resources.
package instance

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/l-nmch/provider-incus/config/common"
)

// Configure configures the instance group.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("incus_instance", func(r *config.Resource) {
		r.ShortGroup = "instance"
		r.Kind = "Instance"
		common.Configure(r)
		r.References["profiles"] = config.Reference{
			TerraformName: "incus_profile",
		}
		// Resolved from the image by the server, and conflicts with "image".
		r.LateInitializer.IgnoredFields = append(r.LateInitializer.IgnoredFields, "architecture")
	})
	p.AddResourceConfigurator("incus_instance_snapshot", func(r *config.Resource) {
		r.ShortGroup = "instance"
		r.Kind = "Snapshot"
		common.Configure(r)
		r.References["instance"] = config.Reference{
			TerraformName: "incus_instance",
		}
	})
}
