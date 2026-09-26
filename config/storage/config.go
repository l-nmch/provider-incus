// Package storage configures the Incus storage pool, volume, bucket and
// bucket key resources.
package storage

import (
	"github.com/crossplane/upjet/v2/pkg/config"

	"github.com/l-nmch/provider-incus/config/common"
)

// Configure configures the storage group.
func Configure(p *config.Provider) {
	poolRef := config.Reference{TerraformName: "incus_storage_pool"}

	p.AddResourceConfigurator("incus_storage_pool", func(r *config.Resource) {
		r.ShortGroup = "storage"
		r.Kind = "Pool"
		common.Configure(r)
		common.ConfigurePendingAware(r)
	})
	p.AddResourceConfigurator("incus_storage_volume", func(r *config.Resource) {
		r.ShortGroup = "storage"
		r.Kind = "Volume"
		common.Configure(r)
		r.References["pool"] = poolRef
	})
	p.AddResourceConfigurator("incus_storage_bucket", func(r *config.Resource) {
		r.ShortGroup = "storage"
		r.Kind = "Bucket"
		common.Configure(r)
		r.References["pool"] = poolRef
	})
	p.AddResourceConfigurator("incus_storage_bucket_key", func(r *config.Resource) {
		r.ShortGroup = "storage"
		r.Kind = "BucketKey"
		common.Configure(r)
		r.References["pool"] = poolRef
		r.References["storage_bucket"] = config.Reference{
			TerraformName: "incus_storage_bucket",
		}
		// Expose the S3 credentials under stable keys in the connection
		// secret, next to the generated "attribute.*" ones.
		r.Sensitive.AdditionalConnectionDetailsFn = func(attr map[string]any) (map[string][]byte, error) {
			conn := map[string][]byte{}
			for _, k := range []string{"access_key", "secret_key"} {
				if v, ok := attr[k].(string); ok {
					conn[k] = []byte(v)
				}
			}
			return conn, nil
		}
	})
}
