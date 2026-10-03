# v0.1.4

- Fix `Image` never syncing when it references a project (or has an external
  name that isn't `<remote>:<fingerprint>`): the Terraform provider crashed on
  every refresh (#1).
- Fix cluster-wide pools and networks with resolved references possibly being
  updated instead of created in an Incus cluster (same root cause as #1).
- Add `imageRef`/`imageSelector` to `Instance`, to create an instance after a
  managed `Image` and from its fingerprint (#2).
- An image can be imported by setting its external name to its fingerprint.

# v0.1.3

- Marketplace README with installation and configuration steps, and a
  dedicated icon, published through the package metadata.

# v0.1.2

- Packaging only, superseded by v0.1.3.

# v0.1.1

First release. All 21 resources of terraform-provider-incus v1.2.0, namespaced
and cluster-scoped, for Crossplane v2.
