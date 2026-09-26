// Package common contains configuration shared by all Incus resource groups.
package common

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure applies the conventions shared by every Incus resource: a
// reference to the owning Project when the resource is project-scoped, and no
// late-initialization of the free-form "config" map, which the server fills
// with defaults and volatile keys that must not be copied back into the spec.
func Configure(r *config.Resource) {
	if _, ok := r.TerraformResource.Schema["project"]; ok && r.Name != "incus_project" {
		r.References["project"] = config.Reference{
			TerraformName: "incus_project",
		}
	}
	if _, ok := r.TerraformResource.Schema["config"]; ok {
		r.LateInitializer.IgnoredFields = append(r.LateInitializer.IgnoredFields, "config")
	}
}

// NestedDefaults sets a default value for an attribute of every element of a
// nested list attribute before the parameters reach Terraform. It fixes
// perpetual diffs on optional, non-computed nested attributes that the
// Terraform provider reads back as an empty string rather than null.
type NestedDefaults struct {
	List  string
	Key   string
	Value any
}

// Convert implements config.TerraformConversion.
func (n NestedDefaults) Convert(params map[string]any, _ *config.Resource, mode config.Mode) (map[string]any, error) {
	if mode != config.ToTerraform {
		return params, nil
	}
	items, _ := params[n.List].([]any)
	for _, i := range items {
		if m, ok := i.(map[string]any); ok {
			if _, set := m[n.Key]; !set {
				m[n.Key] = n.Value
			}
		}
	}
	return params, nil
}
