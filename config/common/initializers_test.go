package common_test

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	ujresource "github.com/crossplane/upjet/v2/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/l-nmch/provider-incus/apis/namespaced/storage/v1alpha1"
	"github.com/l-nmch/provider-incus/config/common"
)

func pool(name, poolName, target string) *v1alpha1.Pool {
	p := &v1alpha1.Pool{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "ns"}}
	p.Spec.ForProvider.Name = ptr.To(poolName)
	p.Spec.ForProvider.Driver = ptr.To("dir")
	if target != "" {
		p.Spec.ForProvider.Target = ptr.To(target)
	}
	return p
}

func created(p *v1alpha1.Pool) *v1alpha1.Pool {
	meta.SetExternalCreateSucceeded(p, metav1.Now().Time)
	p.SetConditions(xpv2.Condition{Type: ujresource.TypeLastAsyncOperation, Status: corev1.ConditionTrue})
	return p
}

func TestPendingAwareInitializer(t *testing.T) {
	cases := map[string]struct {
		mg          *v1alpha1.Pool
		others      []client.Object
		wantSeeded  bool
		wantCreated bool
	}{
		"SingleServer": {
			mg: pool("p", "pool", ""),
		},
		"MemberDefinition": {
			mg:     pool("p-1", "pool", "node1"),
			others: []client.Object{pool("p", "pool", "")},
		},
		"ClusterWideNotCreated": {
			mg:         pool("p", "pool", ""),
			others:     []client.Object{pool("p-1", "pool", "node1")},
			wantSeeded: true,
		},
		"ClusterWideOtherPoolMembers": {
			mg:     pool("p", "pool", ""),
			others: []client.Object{pool("q-1", "other", "node1")},
		},
		"ClusterWideCreated": {
			mg:          created(pool("p", "pool", "")),
			others:      []client.Object{pool("p-1", "pool", "node1")},
			wantCreated: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := runtime.NewScheme()
			if err := v1alpha1.SchemeBuilder.AddToScheme(s); err != nil {
				t.Fatal(err)
			}
			kube := fake.NewClientBuilder().WithScheme(s).WithObjects(append(tc.others, tc.mg)...).Build()
			if err := kube.Get(context.Background(), client.ObjectKeyFromObject(tc.mg), tc.mg); err != nil {
				t.Fatal(err)
			}
			if err := common.PendingAwareInitializer(kube).Initialize(context.Background(), tc.mg); err != nil {
				t.Fatal(err)
			}
			obs, _ := tc.mg.GetObservation()
			if seeded := obs["name"] != nil && obs["name"] != "pool"; seeded != tc.wantSeeded {
				t.Errorf("seeded = %v (name %v), want %v", seeded, obs["name"], tc.wantSeeded)
			}
			stored := &v1alpha1.Pool{}
			if err := kube.Get(context.Background(), client.ObjectKeyFromObject(tc.mg), stored); err != nil {
				t.Fatal(err)
			}
			if _, ok := stored.GetAnnotations()[common.AnnotationKeyCreated]; ok != tc.wantCreated {
				t.Errorf("created annotation persisted = %v, want %v", ok, tc.wantCreated)
			}
		})
	}
}
