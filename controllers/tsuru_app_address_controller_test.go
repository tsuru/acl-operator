package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/acl-operator/api/scheme"
	"github.com/tsuru/acl-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestControllerResolveEmpty(t *testing.T) {
	tsuruAppAddress := &v1alpha1.TsuruAppAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-other-app",
		},
		Spec: v1alpha1.TsuruAppAddressSpec{
			Name: "my-other-app",
		},
		Status: v1alpha1.ResourceAddressStatus{
			IPs:       []string{"10.1.1.57"},
			Pool:      "my-pool",
			Ready:     true,
			UpdatedAt: time.Now().UTC().Add(time.Hour * -1).String(),
		},
	}

	controller := &TsuruAppAddressReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(tsuruAppAddress).Build(),
		Scheme:   scheme.Scheme,
		TsuruAPI: &fakeTsuruAPI{},
		Resolver: &fakeResolver{
			hosts: map[string][]string{
				"myapp.io": {},
			},
		},
	}

	_, err := controller.Reconcile(context.Background(), controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      tsuruAppAddress.Name,
			Namespace: tsuruAppAddress.Namespace,
		},
	})

	require.NoError(t, err)

	existingTsuruAppAddress := &v1alpha1.TsuruAppAddress{}
	err = controller.Get(context.Background(), types.NamespacedName{
		Name:      tsuruAppAddress.Name,
		Namespace: tsuruAppAddress.Namespace,
	}, existingTsuruAppAddress)
	require.NoError(t, err)

	assert.Equal(t, []string{"10.1.1.57"}, existingTsuruAppAddress.Status.IPs)
	assert.False(t, existingTsuruAppAddress.Status.Ready)
	assert.Equal(t, "host myapp.io returned a empty string by resolver", existingTsuruAppAddress.Status.Reason)
}

func TestControllerResolveWithError(t *testing.T) {
	tsuruAppAddress := &v1alpha1.TsuruAppAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-other-app",
		},
		Spec: v1alpha1.TsuruAppAddressSpec{
			Name: "my-other-app",
		},
		Status: v1alpha1.ResourceAddressStatus{
			IPs:       []string{"10.1.1.57"},
			Pool:      "my-pool",
			Ready:     true,
			UpdatedAt: time.Now().UTC().Add(time.Hour * -1).String(),
		},
	}

	controller := &TsuruAppAddressReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(tsuruAppAddress).Build(),
		Scheme:   scheme.Scheme,
		TsuruAPI: &fakeTsuruAPI{},
		Resolver: &fakeResolver{
			errors: map[string]error{
				"myapp.io": errors.New("a error"),
			},
		},
	}

	_, err := controller.Reconcile(context.Background(), controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      tsuruAppAddress.Name,
			Namespace: tsuruAppAddress.Namespace,
		},
	})

	require.NoError(t, err)

	existingTsuruAppAddress := &v1alpha1.TsuruAppAddress{}
	err = controller.Get(context.Background(), types.NamespacedName{
		Name:      tsuruAppAddress.Name,
		Namespace: tsuruAppAddress.Namespace,
	}, existingTsuruAppAddress)
	require.NoError(t, err)

	assert.Equal(t, []string{"10.1.1.57"}, existingTsuruAppAddress.Status.IPs)
	assert.False(t, existingTsuruAppAddress.Status.Ready)
	assert.Equal(t, "a error", existingTsuruAppAddress.Status.Reason)
}
