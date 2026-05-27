package controllers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsuru/acl-operator/api/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestServiceCacheGetByIP_FillsOnFirstCall(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeLoadBalancer,
			ClusterIP: "10.0.0.1",
			Selector: map[string]string{
				"app": "my-svc",
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{IP: "1.2.3.4"},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(svc).Build()
	cache := &serviceCache{Client: client}

	ctx := context.Background()

	result, err := cache.GetByIP(ctx, "1.2.3.4")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "my-svc", result.Name)

	result, err = cache.GetByIP(ctx, "10.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "my-svc", result.Name)

	result, err = cache.GetByIP(ctx, "9.9.9.9")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestServiceCacheGetByIP_UsesCache(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeLoadBalancer,
			ClusterIP: "10.0.0.1",
			Selector:  map[string]string{"app": "my-svc"},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{IP: "1.2.3.4"},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(svc).Build()
	cache := &serviceCache{Client: client}

	ctx := context.Background()

	// First call fills the cache
	result, err := cache.GetByIP(ctx, "1.2.3.4")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify cache is populated and not expired
	expires := cache.allServicesExpires.Load()
	require.NotNil(t, expires)
	assert.True(t, expires.After(time.Now().UTC()), "cache expiry should be in the future")

	// Second call should use cached data (even if we swap client to one with no data)
	cache.Client = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	result, err = cache.GetByIP(ctx, "1.2.3.4")
	require.NoError(t, err)
	require.NotNil(t, result, "should return cached service even though underlying client has no data")
	assert.Equal(t, "my-svc", result.Name)
}

func TestServiceCacheGetByIP_RefreshesAfterExpiry(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeLoadBalancer,
			ClusterIP: "10.0.0.1",
			Selector:  map[string]string{"app": "old-svc"},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{IP: "1.2.3.4"},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(svc).Build()
	cache := &serviceCache{Client: client}

	ctx := context.Background()

	// Fill cache
	result, err := cache.GetByIP(ctx, "1.2.3.4")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "old-svc", result.Name)

	// Force expiry by setting expires to the past
	past := time.Now().UTC().Add(-time.Minute)
	cache.allServicesExpires.Store(&past)

	// Now swap client to one with a new service
	newSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeLoadBalancer,
			ClusterIP: "10.0.0.2",
			Selector:  map[string]string{"app": "new-svc"},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{IP: "5.6.7.8"},
				},
			},
		},
	}
	cache.Client = fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(newSvc).Build()

	// Old IP should be gone, new IP should be found
	result, err = cache.GetByIP(ctx, "1.2.3.4")
	require.NoError(t, err)
	assert.Nil(t, result, "old service should no longer be in cache after refresh")

	result, err = cache.GetByIP(ctx, "5.6.7.8")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "new-svc", result.Name)
}

func TestServiceCacheGetByIP_ClusterIPs(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dual-stack-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type:       corev1.ServiceTypeClusterIP,
			ClusterIP:  "10.0.0.1",
			ClusterIPs: []string{"10.0.0.1", "fd00::1"},
			Selector:   map[string]string{"app": "dual-stack"},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(svc).Build()
	cache := &serviceCache{Client: client}

	ctx := context.Background()

	result, err := cache.GetByIP(ctx, "10.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "dual-stack-svc", result.Name)

	result, err = cache.GetByIP(ctx, "fd00::1")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "dual-stack-svc", result.Name)
}

func TestServiceCacheAtomicPointerInit(t *testing.T) {
	// Ensure zero-value serviceCache works (no panics)
	cache := &serviceCache{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).Build(),
	}

	var ptr atomic.Pointer[mapServiceCache]
	assert.Nil(t, ptr.Load())

	ctx := context.Background()
	result, err := cache.GetByIP(ctx, "1.1.1.1")
	require.NoError(t, err)
	assert.Nil(t, result)
}
