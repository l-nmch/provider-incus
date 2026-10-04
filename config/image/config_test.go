package image

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/l-nmch/provider-incus/apis/cluster/image/v1alpha1"
)

const observedID = ":0eed34fb"

func TestResourceID(t *testing.T) {
	cases := map[string]struct {
		externalName string
		want         string
	}{
		"NotCreatedYet":         {externalName: "", want: placeholderID},
		"ObservedDefaultRemote": {externalName: observedID, want: observedID},
		"ObservedNamedRemote":   {externalName: "crossplane:0eed34fb", want: "crossplane:0eed34fb"},
		"BareFingerprint":       {externalName: "0eed34fb", want: observedID},
		"BareName":              {externalName: "alpine-vm", want: ":alpine-vm"},
		"EmptyFingerprint":      {externalName: "crossplane:", want: placeholderID},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := resourceID(tc.externalName); got != tc.want {
				t.Errorf("resourceID(%q) = %q, want %q", tc.externalName, got, tc.want)
			}
		})
	}
}

func TestResourceIDInitializer(t *testing.T) {
	cases := map[string]struct {
		externalName string
		observed     string
		want         string
	}{
		"NotCreatedYet":    {want: placeholderID},
		"BareExternalName": {externalName: "alpine-vm", want: ":alpine-vm"},
		"Observed":         {externalName: observedID, observed: observedID, want: observedID},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := runtime.NewScheme()
			if err := v1alpha1.SchemeBuilder.AddToScheme(s); err != nil {
				t.Fatal(err)
			}
			img := &v1alpha1.Image{ObjectMeta: metav1.ObjectMeta{Name: "img"}}
			if tc.externalName != "" {
				meta.SetExternalName(img, tc.externalName)
			}
			if tc.observed != "" {
				img.Status.AtProvider.ResourceID = ptr.To(tc.observed)
			}
			kube := fake.NewClientBuilder().WithScheme(s).WithObjects(img).WithStatusSubresource(img).Build()
			if err := resourceIDInitializer(kube).Initialize(context.Background(), img); err != nil {
				t.Fatal(err)
			}
			// The seeded ID must be stored, not only set in memory: resolving
			// references next replaces the object with the stored one.
			stored := &v1alpha1.Image{}
			if err := kube.Get(context.Background(), client.ObjectKeyFromObject(img), stored); err != nil {
				t.Fatal(err)
			}
			if got := ptr.Deref(stored.Status.AtProvider.ResourceID, ""); got != tc.want {
				t.Errorf("stored resource_id = %q, want %q", got, tc.want)
			}
		})
	}
}
