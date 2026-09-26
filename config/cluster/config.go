// Package cluster configures the server-wide Incus resources: cluster groups,
// server configuration and trusted certificates.
package cluster

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/l-nmch/provider-incus/config/common"
)

// Configure configures the cluster group.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("incus_cluster_group", func(r *config.Resource) {
		r.ShortGroup = "cluster"
		r.Kind = "Group"
		common.Configure(r)
	})
	p.AddResourceConfigurator("incus_server", func(r *config.Resource) {
		r.ShortGroup = "cluster"
		r.Kind = "Server"
		common.Configure(r)
	})
	p.AddResourceConfigurator("incus_certificate", func(r *config.Resource) {
		r.ShortGroup = "cluster"
		r.Kind = "Certificate"
		common.Configure(r)
		r.References["projects"] = config.Reference{
			TerraformName: "incus_project",
		}
	})
}
