package common

import (
	"testing"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/google/go-cmp/cmp"
)

const ports = "ports"

func TestNestedDefaults(t *testing.T) {
	n := NestedDefaults{List: ports, Key: "description", Value: ""}
	cases := map[string]struct {
		mode config.Mode
		in   map[string]any
		want map[string]any
	}{
		"FillsMissing": {
			mode: config.ToTerraform,
			in:   map[string]any{ports: []any{map[string]any{"listen_port": "22"}}},
			want: map[string]any{ports: []any{map[string]any{"listen_port": "22", "description": ""}}},
		},
		"KeepsSet": {
			mode: config.ToTerraform,
			in:   map[string]any{ports: []any{map[string]any{"description": "ssh"}}},
			want: map[string]any{ports: []any{map[string]any{"description": "ssh"}}},
		},
		"NoList": {
			mode: config.ToTerraform,
			in:   map[string]any{"network": "n"},
			want: map[string]any{"network": "n"},
		},
		"FromTerraformUntouched": {
			mode: config.FromTerraform,
			in:   map[string]any{ports: []any{map[string]any{}}},
			want: map[string]any{ports: []any{map[string]any{}}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := n.Convert(tc.in, nil, tc.mode)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("-want +got:\n%s", diff)
			}
		})
	}
}
