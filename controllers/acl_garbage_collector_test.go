package controllers

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/acl-operator/api/scheme"
	"github.com/tsuru/acl-operator/api/v1alpha1"
	tsuruv1 "github.com/tsuru/tsuru/provision/kubernetes/pkg/apis/tsuru/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEmptyDryRunLoop(t *testing.T) {
	ctx := context.Background()

	output := &bytes.Buffer{}
	gc := &ACLGarbageCollector{
		Client:       fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects().Build(),
		DryRun:       true,
		DryRunOutput: output,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)
	assert.Equal(t, "", output.String())
}

func TestLoopCleanAppACLDryRun(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
		},
	}

	output := &bytes.Buffer{}
	gc := &ACLGarbageCollector{
		Client:       fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl).Build(),
		DryRun:       true,
		DryRunOutput: output,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	outputString := output.String()
	assert.Contains(t, outputString, "APP ACL is marked to delete default / my-app")
}

func TestLoopIgnoreAppACL(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	output := &bytes.Buffer{}
	gc := &ACLGarbageCollector{
		Client:       fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl, app).Build(),
		DryRun:       true,
		DryRunOutput: output,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	outputString := output.String()
	assert.Equal(t, "", outputString)
}

func TestLoopCleanAppByNamespaceACL(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "old",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	output := &bytes.Buffer{}
	gc := &ACLGarbageCollector{
		Client:       fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl, app).Build(),
		DryRun:       true,
		DryRunOutput: output,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	outputString := output.String()
	assert.Contains(t, outputString, "APP ACL is marked to delete old / my-app")
}

func TestLoopExternalDNSDryRun(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
						Name: "to-keep.example.com",
					},
				},
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	dnsEntry1 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-keep.example.com",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "to-keep.example.com",
		},
	}

	dnsEntry2 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-delete.example.com",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "to-delete.example.com",
		},
	}

	output := &bytes.Buffer{}
	gc := &ACLGarbageCollector{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(
			acl, app, dnsEntry1, dnsEntry2,
		).Build(),
		DryRun:       true,
		DryRunOutput: output,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	outputString := output.String()
	assert.Contains(t, outputString, "dnsEntry is marked to delete to-delete.example.com")
}

func TestLoopTsuruAddressDryRun(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					TsuruApp: "to-keep",
				},
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	tsuruAddress1 := &v1alpha1.TsuruAppAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-keep",
		},
		Spec: v1alpha1.TsuruAppAddressSpec{
			Name: "to-keep",
		},
	}

	tsuruAddress2 := &v1alpha1.TsuruAppAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-delete",
		},
		Spec: v1alpha1.TsuruAppAddressSpec{
			Name: "to-delete",
		},
	}

	output := &bytes.Buffer{}
	gc := &ACLGarbageCollector{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(
			acl, app, tsuruAddress1, tsuruAddress2,
		).Build(),
		DryRun:       true,
		DryRunOutput: output,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	outputString := output.String()
	assert.Contains(t, outputString, "tsuruApp is marked to delete: \"to-delete\"")
}

func TestLoopRPaaSAddressDryRun(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					RpaasInstance: &v1alpha1.ACLSpecRpaasInstance{
						ServiceName: "rpaasv2",
						Instance:    "to-keep",
					},
				},
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	rpaasInstanceAddress1 := &v1alpha1.RpaasInstanceAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-keep",
		},
		Spec: v1alpha1.RpaasInstanceAddressSpec{
			ServiceName: "rpaasv2",
			Instance:    "to-keep",
		},
	}

	rpaasInstanceAddress2 := &v1alpha1.RpaasInstanceAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "rpaasv2-to-delete",
		},
		Spec: v1alpha1.RpaasInstanceAddressSpec{
			ServiceName: "rpaasv2",
			Instance:    "to-delete",
		},
	}

	output := &bytes.Buffer{}
	gc := &ACLGarbageCollector{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(
			acl, app, rpaasInstanceAddress1, rpaasInstanceAddress2,
		).Build(),
		DryRun:       true,
		DryRunOutput: output,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	outputString := output.String()
	assert.Contains(t, outputString, "rpaaInstance is marked to delete rpaasv2-to-delete")
}

func TestLoopExternalDNS(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
						Name: "to-keep.example.com",
					},
				},
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	dnsEntry1 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-keep.example.com",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "to-keep.example.com",
		},
	}

	dnsEntry2 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-delete.example.com",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "to-delete.example.com",
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(
		acl, app, dnsEntry1, dnsEntry2,
	).Build()
	gc := &ACLGarbageCollector{
		Client: client,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	existingDNSEntry := &v1alpha1.ACLDNSEntry{}
	err = client.Get(ctx, types.NamespacedName{
		Name: "to-delete.example.com",
	}, existingDNSEntry)
	assert.True(t, k8sErrors.IsNotFound(err))
}

func TestLoopTsuruAddress(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					TsuruApp: "to-keep",
				},
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	tsuruAddress1 := &v1alpha1.TsuruAppAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-keep",
		},
		Spec: v1alpha1.TsuruAppAddressSpec{
			Name: "to-keep",
		},
	}

	tsuruAddress2 := &v1alpha1.TsuruAppAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "to-delete",
		},
		Spec: v1alpha1.TsuruAppAddressSpec{
			Name: "to-delete",
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(
		acl, app, tsuruAddress1, tsuruAddress2,
	).Build()
	gc := &ACLGarbageCollector{
		Client: client,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	existingTsuruAppAddress := &v1alpha1.TsuruAppAddress{}
	err = client.Get(ctx, types.NamespacedName{
		Name: "to-delete",
	}, existingTsuruAppAddress)
	assert.True(t, k8sErrors.IsNotFound(err))
}

func TestLoopRPaaSAddress(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					RpaasInstance: &v1alpha1.ACLSpecRpaasInstance{
						ServiceName: "rpaasv2",
						Instance:    "to-keep",
					},
				},
			},
		},
	}

	app := &tsuruv1.App{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-app",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "default",
		},
	}

	rpaasInstanceAddress1 := &v1alpha1.RpaasInstanceAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: validResourceName("rpaasv2-to-keep"),
		},
		Spec: v1alpha1.RpaasInstanceAddressSpec{
			ServiceName: "rpaasv2",
			Instance:    "to-keep",
		},
	}

	rpaasInstanceAddress2 := &v1alpha1.RpaasInstanceAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: validResourceName("rpaasv2-to-delete"),
		},
		Spec: v1alpha1.RpaasInstanceAddressSpec{
			ServiceName: "rpaasv2",
			Instance:    "to-delete",
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(
		acl, app, rpaasInstanceAddress1, rpaasInstanceAddress2,
	).Build()
	gc := &ACLGarbageCollector{
		Client: client,
	}
	err := gc.Loop(ctx)

	require.NoError(t, err)

	existingRpaasInstanceAddress := &v1alpha1.RpaasInstanceAddress{}
	err = client.Get(ctx, types.NamespacedName{
		Name: validResourceName("rpaasv2-to-delete"),
	}, existingRpaasInstanceAddress)
	assert.True(t, k8sErrors.IsNotFound(err))
}

func TestLoopCleanAppACL(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      "my-app",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "my-app",
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl).Build()
	gc := &ACLGarbageCollector{
		Client: client,
	}
	err := gc.Loop(ctx)
	require.NoError(t, err)

	existingACL := &v1alpha1.ACL{}
	err = client.Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name:      "my-app",
	}, existingACL)
	assert.True(t, k8sErrors.IsNotFound(err))
}

func TestLoopCleanJobACL(t *testing.T) {
	ctx := context.Background()

	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "default",
			Name:      tsuruJobACLPrefix + "my-job",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruJob: "my-job",
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl).Build()
	gc := &ACLGarbageCollector{
		Client: client,
	}
	err := gc.Loop(ctx)
	require.NoError(t, err)

	existingACL := &v1alpha1.ACL{}
	err = client.Get(ctx, types.NamespacedName{
		Namespace: acl.Namespace,
		Name:      acl.Name,
	}, existingACL)
	assert.True(t, k8sErrors.IsNotFound(err))
}
