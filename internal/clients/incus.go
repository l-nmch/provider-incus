package clients

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/upjet/v2/pkg/terraform"

	clusterv1beta1 "github.com/l-nmch/provider-incus/apis/cluster/v1beta1"
	namespacedv1beta1 "github.com/l-nmch/provider-incus/apis/namespaced/v1beta1"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal incus credentials as JSON"
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		pcSpec, err := resolveProviderConfig(ctx, client, mg)
		if err != nil {
			return terraform.Setup{}, errors.Wrap(err, "cannot resolve provider config")
		}

		data, err := resource.CommonCredentialExtractor(ctx, pcSpec.Credentials.Source, client, pcSpec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		creds := Credentials{}
		if err := json.Unmarshal(data, &creds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}

		cfg, err := creds.terraformConfiguration(os.TempDir())
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		ps.Configuration = cfg
		return ps, nil
	}
}

// remoteName is the name of the Incus remote the Terraform provider is
// configured with. Resources never set "remote", so they all use this default.
const remoteName = "crossplane"

// Credentials is the JSON document expected in the ProviderConfig credentials
// source.
type Credentials struct {
	// Address of the Incus API, e.g. https://incus.example.org:8443.
	Address string `json:"address"`
	// ClientCert and ClientKey are the PEM encoded TLS client certificate and
	// key trusted by the Incus server.
	ClientCert string `json:"client_cert,omitempty"`
	ClientKey  string `json:"client_key,omitempty"`
	// ServerCert is the PEM encoded certificate of the Incus server, needed
	// when it is not signed by a CA trusted by the system.
	ServerCert string `json:"server_cert,omitempty"`
	// Token is a trust token used to have the client certificate trusted on
	// first connection, as an alternative to trusting it out of band.
	Token string `json:"token,omitempty"`
	// AcceptRemoteCertificate accepts the server certificate without
	// verification when ServerCert is not set.
	AcceptRemoteCertificate bool `json:"accept_remote_certificate,omitempty"`
}

// terraformConfiguration writes the credentials as an Incus client
// configuration directory under baseDir, which is the only way the Terraform
// provider accepts TLS client credentials, and returns the matching provider
// configuration. The directory is named after a hash of the credentials so
// that concurrent reconciles share it and credential rotations get a new one.
func (c Credentials) terraformConfiguration(baseDir string) (map[string]any, error) {
	if c.Address == "" {
		return nil, errors.New("address is required")
	}
	if (c.ClientCert == "") != (c.ClientKey == "") {
		return nil, errors.New("client_cert and client_key must be set together")
	}

	raw, err := json.Marshal(c) //nolint:gosec // only hashed to name the directory, never stored
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(raw)
	dir := filepath.Join(baseDir, "provider-incus-"+hex.EncodeToString(sum[:8]))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, errors.Wrap(err, "cannot create incus config directory")
	}

	files := map[string]string{
		"client.crt": c.ClientCert,
		"client.key": c.ClientKey,
		filepath.Join("servercerts", remoteName+".crt"): c.ServerCert,
	}
	for name, content := range files {
		if content == "" {
			continue
		}
		if err := writeFileAtomic(filepath.Join(dir, name), []byte(content)); err != nil {
			return nil, err
		}
	}

	remote := map[string]any{
		"name":                remoteName,
		"address":             c.Address,
		"protocol":            "incus",
		"authentication_type": "tls",
	}
	if c.Token != "" {
		remote["token"] = c.Token
	}
	return map[string]any{
		"config_dir":                   dir,
		"default_remote":               remoteName,
		"generate_client_certificates": c.ClientCert == "",
		"accept_remote_certificate":    c.AcceptRemoteCertificate,
		"remote":                       []any{remote},
	}, nil
}

// writeFileAtomic writes content to path with owner-only permissions, unless
// the file already holds exactly that content.
func writeFileAtomic(path string, content []byte) error {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, content) { //nolint:gosec // path is built from a hash
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return errors.Wrap(err, "cannot create incus config directory")
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return errors.Wrap(err, "cannot create temporary file")
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck // best effort cleanup, renamed on success
	if _, err := tmp.Write(content); err != nil {
		tmp.Close() //nolint:errcheck,gosec // already failing
		return errors.Wrap(err, "cannot write temporary file")
	}
	if err := tmp.Close(); err != nil {
		return errors.Wrap(err, "cannot close temporary file")
	}
	return errors.Wrap(os.Rename(tmp.Name(), path), "cannot write incus config file")
}

func toSharedPCSpec(pc *clusterv1beta1.ProviderConfig) (*namespacedv1beta1.ProviderConfigSpec, error) {
	if pc == nil {
		return nil, nil
	}
	data, err := json.Marshal(pc.Spec)
	if err != nil {
		return nil, err
	}

	var mSpec namespacedv1beta1.ProviderConfigSpec
	err = json.Unmarshal(data, &mSpec)
	return &mSpec, err
}

func resolveProviderConfig(ctx context.Context, crClient client.Client, mg resource.Managed) (*namespacedv1beta1.ProviderConfigSpec, error) {
	switch managed := mg.(type) {
	case resource.LegacyManaged: //nolint:staticcheck // still handling cluster-scoped behavior
		return resolveLegacy(ctx, crClient, managed)
	case resource.ModernManaged:
		return resolveModern(ctx, crClient, managed)
	default:
		return nil, errors.New("resource is not a managed resource")
	}
}

func resolveLegacy(ctx context.Context, client client.Client, mg resource.LegacyManaged) (*namespacedv1beta1.ProviderConfigSpec, error) { //nolint:staticcheck // still handling cluster-scoped behavior
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}
	pc := &clusterv1beta1.ProviderConfig{}
	if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	t := resource.NewLegacyProviderConfigUsageTracker(client, &clusterv1beta1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	return toSharedPCSpec(pc)
}

func resolveModern(ctx context.Context, crClient client.Client, mg resource.ModernManaged) (*namespacedv1beta1.ProviderConfigSpec, error) {
	configRef := mg.GetProviderConfigReference()
	if configRef == nil {
		return nil, errors.New(errNoProviderConfig)
	}

	pcRuntimeObj, err := crClient.Scheme().New(namespacedv1beta1.SchemeGroupVersion.WithKind(configRef.Kind))
	if err != nil {
		return nil, errors.Wrap(err, "unknown GVK for ProviderConfig")
	}
	pcObj, ok := pcRuntimeObj.(client.Object)
	if !ok {
		// This indicates a programming error, types are not properly generated
		return nil, errors.New(" is not an Object")
	}

	// Namespace will be ignored if the PC is a cluster-scoped type
	if err := crClient.Get(ctx, types.NamespacedName{Name: configRef.Name, Namespace: mg.GetNamespace()}, pcObj); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	var pcSpec namespacedv1beta1.ProviderConfigSpec
	pcu := &namespacedv1beta1.ProviderConfigUsage{}
	switch pc := pcObj.(type) {
	case *namespacedv1beta1.ProviderConfig:
		pcSpec = pc.Spec
		if pcSpec.Credentials.SecretRef != nil {
			pcSpec.Credentials.SecretRef.Namespace = mg.GetNamespace()
		}
	case *namespacedv1beta1.ClusterProviderConfig:
		pcSpec = pc.Spec
	default:
		return nil, errors.New("unknown provider config type")
	}
	t := resource.NewProviderConfigUsageTracker(crClient, pcu)
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}
	return &pcSpec, nil
}
