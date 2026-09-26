// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	certificate "github.com/l-nmch/provider-incus/internal/controller/namespaced/cluster/certificate"
	group "github.com/l-nmch/provider-incus/internal/controller/namespaced/cluster/group"
	server "github.com/l-nmch/provider-incus/internal/controller/namespaced/cluster/server"
	image "github.com/l-nmch/provider-incus/internal/controller/namespaced/image/image"
	instance "github.com/l-nmch/provider-incus/internal/controller/namespaced/instance/instance"
	snapshot "github.com/l-nmch/provider-incus/internal/controller/namespaced/instance/snapshot"
	acl "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/acl"
	addressset "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/addressset"
	forward "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/forward"
	integration "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/integration"
	loadbalancer "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/loadbalancer"
	network "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/network"
	peer "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/peer"
	zone "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/zone"
	zonerecord "github.com/l-nmch/provider-incus/internal/controller/namespaced/network/zonerecord"
	profile "github.com/l-nmch/provider-incus/internal/controller/namespaced/profile/profile"
	project "github.com/l-nmch/provider-incus/internal/controller/namespaced/project/project"
	providerconfig "github.com/l-nmch/provider-incus/internal/controller/namespaced/providerconfig"
	bucket "github.com/l-nmch/provider-incus/internal/controller/namespaced/storage/bucket"
	bucketkey "github.com/l-nmch/provider-incus/internal/controller/namespaced/storage/bucketkey"
	pool "github.com/l-nmch/provider-incus/internal/controller/namespaced/storage/pool"
	volume "github.com/l-nmch/provider-incus/internal/controller/namespaced/storage/volume"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		certificate.Setup,
		group.Setup,
		server.Setup,
		image.Setup,
		instance.Setup,
		snapshot.Setup,
		acl.Setup,
		addressset.Setup,
		forward.Setup,
		integration.Setup,
		loadbalancer.Setup,
		network.Setup,
		peer.Setup,
		zone.Setup,
		zonerecord.Setup,
		profile.Setup,
		project.Setup,
		providerconfig.Setup,
		bucket.Setup,
		bucketkey.Setup,
		pool.Setup,
		volume.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		certificate.SetupGated,
		group.SetupGated,
		server.SetupGated,
		image.SetupGated,
		instance.SetupGated,
		snapshot.SetupGated,
		acl.SetupGated,
		addressset.SetupGated,
		forward.SetupGated,
		integration.SetupGated,
		loadbalancer.SetupGated,
		network.SetupGated,
		peer.SetupGated,
		zone.SetupGated,
		zonerecord.SetupGated,
		profile.SetupGated,
		project.SetupGated,
		providerconfig.SetupGated,
		bucket.SetupGated,
		bucketkey.SetupGated,
		pool.SetupGated,
		volume.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		certificate.SetupWebhookWithManager,
		group.SetupWebhookWithManager,
		server.SetupWebhookWithManager,
		image.SetupWebhookWithManager,
		instance.SetupWebhookWithManager,
		snapshot.SetupWebhookWithManager,
		acl.SetupWebhookWithManager,
		addressset.SetupWebhookWithManager,
		forward.SetupWebhookWithManager,
		integration.SetupWebhookWithManager,
		loadbalancer.SetupWebhookWithManager,
		network.SetupWebhookWithManager,
		peer.SetupWebhookWithManager,
		zone.SetupWebhookWithManager,
		zonerecord.SetupWebhookWithManager,
		profile.SetupWebhookWithManager,
		project.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
		bucket.SetupWebhookWithManager,
		bucketkey.SetupWebhookWithManager,
		pool.SetupWebhookWithManager,
		volume.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
