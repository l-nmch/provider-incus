# Changelog

## [0.1.6](https://github.com/l-nmch/provider-incus/compare/v0.1.5...v0.1.6) (2026-10-27)


### Bug Fixes

* **deps:** update all non-major go dependencies ([8964293](https://github.com/l-nmch/provider-incus/commit/8964293ba40d88a843e9169b63509fbfe29079bb))
* **deps:** update all non-major go dependencies ([044460d](https://github.com/l-nmch/provider-incus/commit/044460d0dc9f8307e6f5d9f2c9493b6cae6289e3))
* **deps:** update k8s.io/utils digest to ca04ba5 ([001ec63](https://github.com/l-nmch/provider-incus/commit/001ec635f6f83f86d6dc3f629c10b65a2abdae09))
* **deps:** update k8s.io/utils digest to ca04ba5 ([a3bfed4](https://github.com/l-nmch/provider-incus/commit/a3bfed4f853f73daf1f802105932f871fd6157e6))

## 0.1.5

- Update terraform-provider-incus to v1.2.1, built from source with Go 1.26.10
  and golang.org/x/net v0.60.0 instead of using the released binary, which
  carries the vulnerable Go standard library and golang.org/x/net.
- Build with Go 1.26.10 and golang.org/x/net v0.60.0, which fix the Go
  standard library and HTTP/2 vulnerabilities reported by govulncheck.
- Update Upjet to v2.5.1, which fixes null and empty collections being treated
  as drift, and crossplane-runtime to the stable v2.4.3.

## 0.1.4

- Fix `Image` never syncing when it references a project (or has an external
  name that isn't `<remote>:<fingerprint>`): the Terraform provider crashed on
  every refresh (#1).
- Fix cluster-wide pools and networks with resolved references possibly being
  updated instead of created in an Incus cluster (same root cause as #1).
- Add `imageRef`/`imageSelector` to `Instance`, to create an instance after a
  managed `Image` and from its fingerprint (#2).
- An image can be imported by setting its external name to its fingerprint.

## 0.1.3

- Marketplace README with installation and configuration steps, and a
  dedicated icon, published through the package metadata.

## 0.1.2

- Packaging only, superseded by v0.1.3.

## 0.1.1

First release. All 21 resources of terraform-provider-incus v1.2.0, namespaced
and cluster-scoped, for Crossplane v2.
