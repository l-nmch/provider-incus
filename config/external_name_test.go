package config

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const name = "name"

func TestExternalNameFromState(t *testing.T) {
	cases := map[string]struct {
		resource string
		tfstate  map[string]any
		want     string
		wantErr  bool
	}{
		"Name":          {resource: "incus_instance", tfstate: map[string]any{name: "c1"}, want: "c1"},
		"ListenAddress": {resource: "incus_network_forward", tfstate: map[string]any{"listen_address": "10.0.0.1"}, want: "10.0.0.1"},
		"ZoneName":      {resource: "incus_network_zone", tfstate: map[string]any{name: "example.org"}, want: "example.org"},
		"Image":         {resource: "incus_image", tfstate: map[string]any{"resource_id": "r:fp"}, want: "r:fp"},
		"ServerTarget":  {resource: "incus_server", tfstate: map[string]any{"target": "node1"}, want: "node1"},
		"ServerGlobal":  {resource: "incus_server", tfstate: map[string]any{}, want: "server"},
		"Missing":       {resource: "incus_profile", tfstate: map[string]any{}, wantErr: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ExternalNameConfigs[tc.resource].GetExternalNameFn(tc.tfstate)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExternalNameSetsArguments(t *testing.T) {
	cases := map[string]struct {
		resource string
		params   map[string]any
		extName  string
		want     map[string]any
	}{
		"NameFromExternalName": {
			resource: "incus_profile", params: map[string]any{}, extName: "p1",
			want: map[string]any{name: "p1"},
		},
		"VolumeTypeDefault": {
			resource: "incus_storage_volume", params: map[string]any{}, extName: "v1",
			want: map[string]any{name: "v1", "type": "custom"},
		},
		"VolumeTypeKept": {
			resource: "incus_storage_volume", params: map[string]any{"type": "virtual-machine"}, extName: "v1",
			want: map[string]any{name: "v1", "type": "virtual-machine"},
		},
		"SpecNameUntouched": {
			resource: "incus_network", params: map[string]any{name: "net"}, extName: "other",
			want: map[string]any{name: "net"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ExternalNameConfigs[tc.resource].SetIdentifierArgumentFn(tc.params, tc.extName)
			if diff := cmp.Diff(tc.want, tc.params); diff != "" {
				t.Errorf("-want +got:\n%s", diff)
			}
			id, err := ExternalNameConfigs[tc.resource].GetIDFn(context.Background(), tc.extName, tc.params, nil)
			if err != nil || id != tc.extName {
				t.Errorf("GetIDFn = %q, %v", id, err)
			}
		})
	}
}
