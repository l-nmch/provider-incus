package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/pkg/errors"
)

// ExternalNameConfigs contains all external name configurations for this
// provider. None of the Incus Terraform resources expose an "id" attribute:
// they are identified by their name (scoped by project/pool/network through
// regular parameters), except network forwards and load balancers which are
// identified by their listen address, images which are identified by a
// server-assigned fingerprint and the server singleton.
var ExternalNameConfigs = map[string]config.ExternalName{
	"incus_certificate":         nameAsIdentifier(),
	"incus_cluster_group":       nameAsIdentifier(),
	"incus_image":               specIdentifier("resource_id"),
	"incus_instance":            nameAsIdentifier(),
	"incus_instance_snapshot":   nameAsIdentifier(),
	"incus_network":             specIdentifier("name"),
	"incus_network_acl":         nameAsIdentifier(),
	"incus_network_address_set": nameAsIdentifier(),
	"incus_network_forward":     specIdentifier("listen_address"),
	"incus_network_integration": nameAsIdentifier(),
	"incus_network_lb":          specIdentifier("listen_address"),
	"incus_network_peer":        nameAsIdentifier(),
	"incus_network_zone":        specIdentifier("name"),
	"incus_network_zone_record": nameAsIdentifier(),
	"incus_profile":             nameAsIdentifier(),
	"incus_project":             nameAsIdentifier(),
	"incus_server":              serverSingleton(),
	"incus_storage_bucket":      nameAsIdentifier(),
	"incus_storage_bucket_key":  nameAsIdentifier(),
	"incus_storage_pool":        specIdentifier("name"),
	"incus_storage_volume":      withDefaults(nameAsIdentifier(), map[string]any{"type": "custom"}),
}

// nameAsIdentifier fills the "name" argument from the external name like
// config.NameAsIdentifier, but reads the external name back from that
// argument in the Terraform state rather than from the "id" attribute, which
// Incus resources don't have.
func nameAsIdentifier() config.ExternalName {
	return fromState(config.NameAsIdentifier, "name")
}

// specIdentifier identifies the resource by an argument that stays a regular
// spec field because metadata.name can't hold it: listen addresses, DNS zone
// names whose dots are invalid in Terraform resource block names, or pools and
// networks which, in a cluster, are defined by several managed resources (one
// per member plus a cluster-wide one) sharing the same name.
func specIdentifier(param string) config.ExternalName {
	return fromState(config.IdentifierFromProvider, param)
}

func fromState(base config.ExternalName, param string) config.ExternalName {
	return config.NewExternalNameFrom(base,
		config.WithGetExternalNameFn(func(_ config.GetExternalNameFn, tfstate map[string]any) (string, error) {
			if v, ok := tfstate[param].(string); ok && v != "" {
				return v, nil
			}
			return "", errors.Errorf("cannot find %q in tfstate", param)
		}),
	)
}

// withDefaults sets the schema defaults of the arguments the Terraform
// provider reads a resource by. Without an "id", Terraform refreshes a resource
// that doesn't exist yet from a state built out of its parameters, where
// defaults of optional+computed arguments are not applied yet.
func withDefaults(e config.ExternalName, defaults map[string]any) config.ExternalName {
	return config.NewExternalNameFrom(e,
		config.WithSetIdentifierArgumentsFn(func(fn config.SetIdentifierArgumentsFn, base map[string]any, externalName string) {
			fn(base, externalName)
			for k, v := range defaults {
				if _, ok := base[k]; !ok {
					base[k] = v
				}
			}
		}),
	)
}

// serverSingleton identifies the server configuration, which always exists
// and only manages the config keys it declares, by its cluster member target
// (or "server" when it applies cluster-wide).
func serverSingleton() config.ExternalName {
	return config.NewExternalNameFrom(config.IdentifierFromProvider,
		config.WithGetExternalNameFn(func(_ config.GetExternalNameFn, tfstate map[string]any) (string, error) {
			if t, ok := tfstate["target"].(string); ok && t != "" {
				return t, nil
			}
			return "server", nil
		}),
	)
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
