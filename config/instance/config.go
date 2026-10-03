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
		// A managed Image is referenced by fingerprint, which Incus accepts
		// in place of an alias and which is only known once the image
		// exists, so the instance waits for it instead of racing it.
		r.References["image"] = config.Reference{
			TerraformName: "incus_image",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("fingerprint", true)`,
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
