# provider-incus

`provider-incus` is a [Crossplane](https://crossplane.io/) provider for
[Incus](https://linuxcontainers.org/incus/), generated with
[Upjet](https://github.com/crossplane/upjet) from
[`lxc/terraform-provider-incus`](https://github.com/lxc/terraform-provider-incus)
**v1.2.0**. It targets Crossplane v2 and exposes every resource of the
Terraform provider, both as namespaced (`*.incus.m.crossplane.io`) and
cluster-scoped (`*.incus.crossplane.io`) managed resources.

## Resources

| Group | Kind | Terraform resource | Identified by |
|---|---|---|---|
| `project` | `Project` | `incus_project` | `metadata.name` |
| `profile` | `Profile` | `incus_profile` | `metadata.name` |
| `instance` | `Instance` | `incus_instance` | `metadata.name` |
| `instance` | `Snapshot` | `incus_instance_snapshot` | `metadata.name` |
| `image` | `Image` | `incus_image` | fingerprint, assigned by the server |
| `cluster` | `Group` | `incus_cluster_group` | `metadata.name` |
| `cluster` | `Certificate` | `incus_certificate` | `metadata.name` |
| `cluster` | `Server` | `incus_server` | singleton (per `target`) |
| `storage` | `Pool` | `incus_storage_pool` | `spec.forProvider.name` |
| `storage` | `Volume` | `incus_storage_volume` | `metadata.name` |
| `storage` | `Bucket` | `incus_storage_bucket` | `metadata.name` |
| `storage` | `BucketKey` | `incus_storage_bucket_key` | `metadata.name` |
| `network` | `Network` | `incus_network` | `spec.forProvider.name` |
| `network` | `ACL` | `incus_network_acl` | `metadata.name` |
| `network` | `AddressSet` | `incus_network_address_set` | `metadata.name` |
| `network` | `Forward` | `incus_network_forward` | `spec.forProvider.listenAddress` |
| `network` | `LoadBalancer` | `incus_network_lb` | `spec.forProvider.listenAddress` |
| `network` | `Peer` | `incus_network_peer` | `metadata.name` |
| `network` | `Integration` | `incus_network_integration` | `metadata.name` |
| `network` | `Zone` | `incus_network_zone` | `spec.forProvider.name` |
| `network` | `ZoneRecord` | `incus_network_zone_record` | `metadata.name` |

Resources identified by `metadata.name` use it as the Incus name unless the
`crossplane.io/external-name` annotation says otherwise. An image's external
name is set to `<remote>:<fingerprint>` once it exists; set it to a bare
fingerprint to import an existing image. Pools and networks
keep their name in the spec because a cluster needs several managed resources
with the same Incus name (see below); zones do because DNS names contain dots,
which Terraform doesn't allow in resource names.

Project-scoped resources accept `projectRef`/`projectSelector`; other
cross-resource references: `imageRef` and `profilesRefs` (Instance),
`instanceRef`
(Snapshot), `poolRef` (Volume, Bucket, BucketKey), `storageBucketRef`
(BucketKey), `networkRef` (Forward, LoadBalancer, Peer), `targetNetworkRef` /
`targetIntegrationRef` (Peer), `zoneRef` (ZoneRecord), `projectsRefs`
(Certificate).

An Instance's `imageRef`/`imageSelector` points at a managed `Image` and
resolves to its fingerprint once the image exists, so the instance is created
after the image instead of failing until it shows up. `image` still accepts any
image Incus resolves, such as `images:alpine/3.22` or a local alias.

## Configuration

Create a Secret holding a JSON document, and a `ProviderConfig` (or
`ClusterProviderConfig`) pointing to it:

```json
{
  "address": "https://incus.example.org:8443",
  "client_cert": "-----BEGIN CERTIFICATE-----\n...",
  "client_key": "-----BEGIN EC PRIVATE KEY-----\n...",
  "server_cert": "-----BEGIN CERTIFICATE-----\n..."
}
```

| Key | Description |
|---|---|
| `address` | Incus API URL (required). |
| `client_cert`, `client_key` | PEM TLS client certificate and key, trusted by the server (`incus config trust add-certificate`). |
| `server_cert` | PEM server certificate to pin, when it isn't signed by a CA the provider trusts. |
| `token` | Trust token, as an alternative to trusting the client certificate out of band (a client certificate is then generated). |
| `accept_remote_certificate` | Accept the server certificate without verification when `server_cert` isn't set. |

```yaml
apiVersion: incus.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
  namespace: my-namespace
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: my-namespace
      name: incus-creds
      key: credentials
```

Restrict what a ProviderConfig can do with a restricted certificate
(`incus config trust add-certificate --restricted --projects ...`).

## Usage notes

- **Clustered pools and networks.** In an Incus cluster, a storage pool (or a
  non-OVN network) is first defined on every member, then created
  cluster-wide. Declare one managed resource per member with `target` set and
  one without `target`, all with the same `spec.forProvider.name`; the
  cluster-wide one is created once the member definitions exist. See
  [`examples/e2e/12-global.yaml`](examples/e2e/12-global.yaml). Once created,
  the cluster-wide resource is annotated `incus.crossplane.io/created`. To adopt an
  existing cluster-wide pool or network alongside member definitions, use
  management policies without `Create`.
- **Bucket keys** publish `access_key` and `secret_key` in the connection
  secret (`writeConnectionSecretToRef`), e.g. to hand a scoped S3 key to a
  workload. Buckets need a pool that supports them (`cephobject`, or local
  buckets served through `core.storage_buckets_address`).
- **Server** only manages the config keys it declares; deleting it unsets
  them.
- Instance `architecture` and every free-form `config` map are not
  late-initialized, so server-side defaults and `volatile.*` keys never end up
  in the spec.

## Developing

Tools: Go (see `go.mod`), `goimports` on `PATH`, Docker for images.

```console
make generate        # schema, docs, CRDs, controllers
make reviewable      # generate + lint + tests
make build           # provider image and xpkg under _output/
```

Provider configuration lives in [`config/`](config/): external names in
[`external_name.go`](config/external_name.go), one package per API group, and
shared behaviour in [`config/common`](config/common). Credentials handling is
in [`internal/clients/incus.go`](internal/clients/incus.go).

### Implementation notes

None of the Incus Terraform resources has an `id` attribute. Upjet therefore
refreshes a resource from a Terraform state built from its parameters, which
the provider works around in a few places:

- external names are read back from the identifying attribute in the state
  (`name`, `listen_address`, `resource_id`) instead of `id`;
- storage volumes get `type = "custom"` before the first refresh, since the
  Terraform provider reads a volume by type and only applies that default on
  create;
- images get their read-only `resource_id` seeded before the refresh (the
  external name once known, a placeholder before creation);
- cluster-wide pools and networks are refreshed under an impossible name until
  created, since the pending definition would otherwise be found and updated
  instead of created;
- these seeded values are stored in the status rather than only set in memory:
  on the first reconcile, resolving references patches the object and replaces
  it with the stored one before upjet writes the Terraform state, which it
  never rewrites afterwards;
- network forward ports get an empty `description` when unset, since the
  Terraform provider reads it back as `""` and would otherwise taint the
  forward after creation.

### End-to-end tests

The manifests in [`examples/e2e/`](examples/e2e/) cover every resource. All
Incus objects they create are prefixed `xptest`, in a dedicated `xptest`
project and Kubernetes namespace. `${XPTEST_CERT}` (a PEM certificate indented
by 6 spaces), `${XPTEST_FWD_ADDR}` and `${XPTEST_LB_ADDR}` (the
`volatile.network.ipv4.address` of `xptest-net` and `xptest-net2`) are filled
with `envsubst`.

1. `hack/e2e/snapshot.sh > before.json` records the Incus cluster state
   (server config values are hashed, never stored).
2. Trust a temporary client certificate, apply `package/crds`, the credentials
   Secret and the ProviderConfig, and run the provider out of cluster with
   `--certs-dir ""`. Set `PLUGIN_UNIX_SOCKET_DIR` to a short directory when
   `TMPDIR` is long: Terraform plugins listen on a UNIX socket whose path is
   limited to 108 characters. Use a provider filesystem mirror
   (`TF_CLI_CONFIG_FILE`, as in the image) rather than a shared plugin cache,
   which concurrent `terraform init` runs corrupt.
3. Apply the manifests in order, check `Synced`/`Ready`, then delete them.
4. `hack/e2e/cleanup.sh [kube-context] [ssh-target]` idempotently removes
   whatever is left on both sides (including the temporary certificate), and
   `hack/e2e/snapshot.sh | diff before.json -` must come out empty.

## License

This provider is licensed under the [Apache License 2.0](LICENSE).

Its container image also ships third-party binaries, redistributed unmodified
under their own license, whose source code is available upstream:

| Component | Version | License | Source |
|---|---|---|---|
| Terraform CLI | 1.5.7 | MPL-2.0 | https://github.com/hashicorp/terraform/tree/v1.5.7 |
| terraform-provider-incus | 1.2.0 | MPL-2.0 | https://github.com/lxc/terraform-provider-incus/tree/v1.2.0 |

Resource and field descriptions in the CRDs are derived from the
terraform-provider-incus documentation (MPL-2.0).
