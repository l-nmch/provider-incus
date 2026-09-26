package common

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/crossplane/upjet/v2/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// notCreatedName is a name no Incus object can have (it is longer than the
// 63 characters Incus allows), so Terraform reads it as not found.
const notCreatedName = "crossplane-not-created-yet-placeholder-that-no-incus-object-can-ever-have"

// SeedObservation sets the given Terraform attributes in the in-memory
// observation when they are unset, before upjet builds the Terraform state
// from the parameters and the observation. It is never persisted: the managed
// reconciler hands this same object to Connect, and the real attributes are
// observed once Terraform has refreshed or created the resource.
func SeedObservation(mg xpresource.Managed, attrs map[string]any) error {
	tr, ok := mg.(resource.Terraformed)
	if !ok {
		return errors.New("managed resource is not Terraformed")
	}
	obs, err := tr.GetObservation()
	if err != nil {
		return errors.Wrap(err, "cannot get observation")
	}
	for k, v := range attrs {
		if s, _ := obs[k].(string); s == "" {
			obs[k] = v
		}
	}
	return errors.Wrap(tr.SetObservation(obs), "cannot set observation")
}

// AnnotationKeyCreated marks cluster-wide pools and networks whose creation
// by Crossplane has completed.
const AnnotationKeyCreated = "incus.crossplane.io/created"

// PendingAwareInitializer handles cluster-wide pools and networks, which are
// first defined on each cluster member by managed resources with the same name
// and a "target", and then created cluster-wide by one without "target".
// Reading the cluster-wide object by name finds it as soon as the per-member
// definitions exist, in a pending state that can't be updated. Terraform never
// reads it before creating it since it has no state for it yet, but upjet
// builds one from the parameters. So until its creation has completed, it is
// read under a name that can't exist, which makes Terraform create it.
// Completion is recorded in an annotation because neither the observation
// (filled from the pending object) nor the external-create-succeeded
// annotation (set before the asynchronous apply finishes) tell it.
func PendingAwareInitializer(kube client.Client) managed.Initializer {
	return managed.InitializerFn(func(ctx context.Context, mg xpresource.Managed) error {
		if _, ok := mg.GetAnnotations()[AnnotationKeyCreated]; ok {
			return nil
		}
		tr, ok := mg.(resource.Terraformed)
		if !ok {
			return errors.New("managed resource is not Terraformed")
		}
		name, err := clusterWideName(tr)
		if err != nil || name == "" {
			return err
		}
		pending, err := hasMemberDefinitions(ctx, kube, mg, name)
		if err != nil || !pending {
			return err
		}
		last := mg.GetCondition(resource.TypeLastAsyncOperation)
		if !meta.GetExternalCreateSucceeded(mg).IsZero() && last.Status == corev1.ConditionTrue {
			meta.AddAnnotations(mg, map[string]string{AnnotationKeyCreated: "true"})
			return errors.Wrap(kube.Update(ctx, mg), "cannot record creation")
		}
		obs, err := tr.GetObservation()
		if err != nil {
			return errors.Wrap(err, "cannot get observation")
		}
		obs["name"] = notCreatedName
		return errors.Wrap(tr.SetObservation(obs), "cannot set observation")
	})
}

// clusterWideName returns the name of the object when the managed resource
// defines it cluster-wide, i.e. without targeting a cluster member.
func clusterWideName(tr resource.Terraformed) (string, error) {
	params, err := tr.GetParameters()
	if err != nil {
		return "", errors.Wrap(err, "cannot get parameters")
	}
	if t, _ := params["target"].(string); t != "" {
		return "", nil
	}
	name, _ := params["name"].(string)
	return name, nil
}

// hasMemberDefinitions tells whether managed resources of the same kind and in
// the same namespace define the named object on a cluster member.
func hasMemberDefinitions(ctx context.Context, kube client.Client, mg xpresource.Managed, name string) (bool, error) {
	gvk, err := apiutil.GVKForObject(mg, kube.Scheme())
	if err != nil {
		return false, errors.Wrap(err, "cannot get kind")
	}
	l := &unstructured.UnstructuredList{}
	l.SetGroupVersionKind(gvk.GroupVersion().WithKind(gvk.Kind + "List"))
	if err := kube.List(ctx, l, client.InNamespace(mg.GetNamespace())); err != nil {
		return false, errors.Wrap(err, "cannot list managed resources")
	}
	for _, o := range l.Items {
		n, _, _ := unstructured.NestedString(o.Object, "spec", "forProvider", "name")
		t, _, _ := unstructured.NestedString(o.Object, "spec", "forProvider", "target")
		if n == name && t != "" {
			return true, nil
		}
	}
	return false, nil
}

// ConfigurePendingAware adds PendingAwareInitializer to a resource.
func ConfigurePendingAware(r *config.Resource) {
	r.InitializerFns = append(r.InitializerFns, PendingAwareInitializer)
}
