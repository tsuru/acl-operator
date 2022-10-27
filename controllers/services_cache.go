package controllers

import (
	"context"
	"time"

	"sync/atomic"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mapServiceCache map[string]*corev1.Service

type serviceCache struct {
	client.Client

	allServices        atomic.Pointer[mapServiceCache]
	allServicesExpires atomic.Pointer[time.Time]
}

func (s *serviceCache) GetByIP(ctx context.Context, ip string) (*corev1.Service, error) {
	allServices := s.allServices.Load()
	expires := s.allServicesExpires.Load()

	if allServices == nil || expires == nil || expires.After(time.Now().UTC()) {
		var err error
		allServices, err = s.fillCache(ctx)
		if err != nil {
			return nil, err
		}
	}

	return (*allServices)[ip], nil
}

func (s *serviceCache) fillCache(ctx context.Context) (*mapServiceCache, error) {
	allServices := corev1.ServiceList{}

	err := s.Client.List(ctx, &allServices, &client.ListOptions{Namespace: metav1.NamespaceAll})
	if err != nil {
		return nil, err
	}

	cache := mapServiceCache{}

	for i, service := range allServices.Items {
		if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
			continue
		}
		if len(service.Status.LoadBalancer.Ingress) == 0 {
			continue
		}

		if service.Status.LoadBalancer.Ingress[0].IP == "" {
			continue
		}

		cache[service.Status.LoadBalancer.Ingress[0].IP] = &allServices.Items[i]
	}

	s.allServices.Store(&cache)
	expires := time.Now().UTC().Add(time.Minute * 15)
	s.allServicesExpires.Store(&expires)

	return &cache, err
}
