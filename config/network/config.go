// Package network configures the Incus network resources.
package network

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/l-nmch/provider-incus/config/common"
)

var kinds = map[string]string{
	"incus_network":             "Network",
	"incus_network_acl":         "ACL",
	"incus_network_address_set": "AddressSet",
	"incus_network_forward":     "Forward",
	"incus_network_integration": "Integration",
	"incus_network_lb":          "LoadBalancer",
	"incus_network_peer":        "Peer",
	"incus_network_zone":        "Zone",
	"incus_network_zone_record": "ZoneRecord",
}

// Configure configures the network group.
func Configure(p *config.Provider) {
	networkRef := config.Reference{TerraformName: "incus_network"}

	for name, kind := range kinds {
		p.AddResourceConfigurator(name, func(r *config.Resource) {
			r.ShortGroup = "network"
			r.Kind = kind
			common.Configure(r)
			if _, ok := r.TerraformResource.Schema["network"]; ok {
				r.References["network"] = networkRef
			}
		})
	}
	p.AddResourceConfigurator("incus_network", common.ConfigurePendingAware)
	p.AddResourceConfigurator("incus_network_forward", func(r *config.Resource) {
		r.TerraformConversions = append(r.TerraformConversions,
			common.NestedDefaults{List: "ports", Key: "description", Value: ""})
	})
	p.AddResourceConfigurator("incus_network_peer", func(r *config.Resource) {
		r.References["target_network"] = networkRef
		r.References["target_integration"] = config.Reference{
			TerraformName: "incus_network_integration",
		}
	})
	p.AddResourceConfigurator("incus_network_zone_record", func(r *config.Resource) {
		r.References["zone"] = config.Reference{
			TerraformName: "incus_network_zone",
		}
	})
}
