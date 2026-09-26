package clients

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTerraformConfiguration(t *testing.T) {
	cases := map[string]struct {
		creds     Credentials
		wantErr   bool
		wantFiles map[string]string
		wantGen   bool
		wantToken bool
	}{
		"MissingAddress": {
			creds:   Credentials{ClientCert: "c", ClientKey: "k"},
			wantErr: true,
		},
		"CertWithoutKey": {
			creds:   Credentials{Address: "https://incus:8443", ClientCert: "c"},
			wantErr: true,
		},
		"TLSClient": {
			creds: Credentials{Address: "https://incus:8443", ClientCert: "cert", ClientKey: "key", ServerCert: "srv"},
			wantFiles: map[string]string{
				"client.crt":                 "cert",
				"client.key":                 "key",
				"servercerts/crossplane.crt": "srv",
			},
		},
		"TrustToken": {
			creds:     Credentials{Address: "https://incus:8443", Token: "tok", AcceptRemoteCertificate: true},
			wantFiles: map[string]string{},
			wantGen:   true,
			wantToken: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			cfg, err := tc.creds.terraformConfiguration(base)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			dir := cfg["config_dir"].(string)
			for name, want := range tc.wantFiles {
				got, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(want, string(got)); diff != "" {
					t.Errorf("%s: -want +got:\n%s", name, diff)
				}
				fi, _ := os.Stat(filepath.Join(dir, name))
				if fi.Mode().Perm() != 0o600 {
					t.Errorf("%s: mode %v, want 0600", name, fi.Mode().Perm())
				}
			}
			if cfg["generate_client_certificates"] != tc.wantGen {
				t.Errorf("generate_client_certificates = %v, want %v", cfg["generate_client_certificates"], tc.wantGen)
			}
			remote := cfg["remote"].([]any)[0].(map[string]any)
			if _, ok := remote["token"]; ok != tc.wantToken {
				t.Errorf("token set = %v, want %v", ok, tc.wantToken)
			}
			// Same credentials must map to the same directory.
			again, err := tc.creds.terraformConfiguration(base)
			if err != nil || again["config_dir"] != dir {
				t.Errorf("config_dir not stable: %v, %v", again["config_dir"], err)
			}
		})
	}
}
