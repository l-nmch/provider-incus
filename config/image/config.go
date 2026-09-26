// Package image configures the Incus image resource.
package image

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/l-nmch/provider-incus/config/common"
)

// placeholderID never matches an image, so reading it tells Terraform the
// image doesn't exist yet.
const placeholderID = "crossplane:unknown"

// Configure configures the image group.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("incus_image", func(r *config.Resource) {
		r.ShortGroup = "image"
		r.Kind = "Image"
		common.Configure(r)
		r.InitializerFns = append(r.InitializerFns, resourceIDInitializer)
	})
}

// resourceIDInitializer seeds the computed resource_id, the only attribute the
// Terraform provider reads an image by and which can't be passed as a
// parameter since it is read-only: the external name once known, or an ID
// that never matches an image so that Terraform creates it.
func resourceIDInitializer(_ client.Client) managed.Initializer {
	return managed.InitializerFn(func(_ context.Context, mg xpresource.Managed) error {
		id := meta.GetExternalName(mg)
		if id == "" {
			id = placeholderID
		}
		return common.SeedObservation(mg, map[string]any{"resource_id": id})
	})
}
