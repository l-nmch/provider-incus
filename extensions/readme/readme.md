# provider-incus

Crossplane provider for [Incus](https://linuxcontainers.org/incus/): manage
projects, instances (containers and VMs), profiles, images, storage, networks
and cluster configuration as Kubernetes resources.

Generated with [Upjet](https://github.com/crossplane/upjet) from
[`lxc/terraform-provider-incus`](https://github.com/lxc/terraform-provider-incus)
v1.2.0. Requires **Crossplane v2**. Source, issues and full documentation:
https://github.com/l-nmch/provider-incus

## Install

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-incus
spec:
  package: xpkg.upbound.io/l-nmch/provider-incus:v0.1.2
```

```console
kubectl apply -f provider.yaml
kubectl get providers.pkg.crossplane.io provider-incus   # wait for HEALTHY=True
```

The provider pods must be able to reach the Incus API (port 8443 by default).

## Configure

### 1. Give the provider access to Incus

Create a TLS client certificate and trust it on the Incus server:

```console
openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:secp384r1 -sha384 -nodes \
  -days 3650 -subj "/CN=crossplane" -keyout client.key -out client.crt
incus config trust add-certificate client.crt --name crossplane
```

Use `--restricted --projects=<p1>,<p2>` to limit the provider to some
projects (it then can't create projects).

### 2. Store the credentials in a Secret

The Secret holds a JSON document:

| Key | Description |
|---|---|
| `address` | Incus API URL (required), e.g. `https://incus.example.org:8443` |
| `client_cert`, `client_key` | PEM client certificate and key trusted by the server |
| `server_cert` | PEM server certificate to pin when it isn't signed by a trusted CA |
| `token` | Trust token, instead of trusting the certificate out of band |
| `accept_remote_certificate` | Skip server certificate verification when `server_cert` is unset |

```console
jq -n --arg crt "$(cat client.crt)" --arg key "$(cat client.key)" \
  '{address: "https://incus.example.org:8443", client_cert: $crt, client_key: $key}' > creds.json
kubectl -n crossplane-system create secret generic incus-creds --from-file=credentials=creds.json
```

### 3. Create a ClusterProviderConfig

```yaml
apiVersion: incus.m.crossplane.io/v1beta1
kind: ClusterProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: incus-creds
      key: credentials
```

Namespaced managed resources use the `ClusterProviderConfig` named `default`
unless they set `providerConfigRef`. A namespaced `ProviderConfig` is also
available.

## Example

```yaml
apiVersion: project.incus.m.crossplane.io/v1alpha1
kind: Project
metadata:
  name: demo
  namespace: default
spec:
  forProvider:
    config:
      features.profiles: "true"
      features.images: "true"
---
apiVersion: instance.incus.m.crossplane.io/v1alpha1
kind: Instance
metadata:
  name: web
  namespace: default
spec:
  forProvider:
    projectRef:
      name: demo
    image: images:alpine/3.22
    type: container
    device:
      - name: root
        type: disk
        properties:
          pool: default
          path: /
      - name: eth0
        type: nic
        properties:
          network: incusbr0
```

```console
kubectl get instances.instance.incus.m.crossplane.io web   # SYNCED/READY=True
```

## Resources

| Group | Kinds |
|---|---|
| `project` | Project |
| `profile` | Profile |
| `instance` | Instance, Snapshot |
| `image` | Image |
| `storage` | Pool, Volume, Bucket, BucketKey |
| `network` | Network, ACL, AddressSet, Forward, LoadBalancer, Peer, Integration, Zone, ZoneRecord |
| `cluster` | Group, Certificate, Server |

Each is available namespaced (`*.incus.m.crossplane.io`) and cluster-scoped
(`*.incus.crossplane.io`).

Good to know:

- Pools, networks and zones take their Incus name from `spec.forProvider.name`;
  forwards and load balancers from `spec.forProvider.listenAddress`; other
  kinds from `metadata.name` (or the `crossplane.io/external-name` annotation).
- In an Incus cluster, a storage pool or a non-OVN network is first defined on
  each member (one resource per member with `target`), then created
  cluster-wide by a resource without `target` sharing the same name.
- `BucketKey` publishes `access_key` and `secret_key` through
  `writeConnectionSecretToRef`.

See the [README](https://github.com/l-nmch/provider-incus#readme) for details.

## License

Apache-2.0. The image also ships Terraform 1.5.7 and terraform-provider-incus
1.2.0, both MPL-2.0. This is a community provider, not affiliated with the
Incus or Crossplane projects.
