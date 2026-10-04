// Package image configures the Incus image resource.
package image

import (
	"context"
	"strings"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/crossplane/upjet/v2/pkg/resource"
	"github.com/pkg/errors"
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
// that never matches an image so that Terraform creates it. An observed
// resource_id is kept as is.
func resourceIDInitializer(kube client.Client) managed.Initializer {
	return managed.InitializerFn(func(ctx context.Context, mg xpresource.Managed) error {
		tr, ok := mg.(resource.Terraformed)
		if !ok {
			return errors.New("managed resource is not Terraformed")
		}
		obs, err := tr.GetObservation()
		if err != nil {
			return errors.Wrap(err, "cannot get observation")
		}
		if id, _ := obs["resource_id"].(string); id != "" && id != placeholderID {
			return nil
		}
		return common.SeedObservation(ctx, kube, mg, map[string]any{"resource_id": resourceID(meta.GetExternalName(mg))})
	})
}

// resourceID turns an external name into a resource_id the Terraform provider
// can read. It expects "<remote>:<fingerprint>" and panics on anything without
// a colon, so a bare value, such as a fingerprint set as external name to
// import an existing image, is taken as a fingerprint on the default remote.
func resourceID(externalName string) string {
	switch {
	case externalName == "":
		return placeholderID
	case !strings.Contains(externalName, ":"):
		return ":" + externalName
	case strings.HasSuffix(externalName, ":"):
		// An empty fingerprint would make the provider list images instead.
		return placeholderID
	default:
		return externalName
	}
}
