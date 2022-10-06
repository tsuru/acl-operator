package controllers

import (
	"context"
	"errors"
	"net"

	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type fakeResolver struct{}

func (f *fakeResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if host == "www.google.com.br" {
		return []net.IPAddr{
			{
				IP: net.ParseIP("8.8.8.8"),
			},
			{
				IP: net.ParseIP("8.8.4.4"),
			},
		}, nil
	}

	return nil, errors.New("no mocks for host")
}

func (suite *ControllerSuite) TestACLDNSEntryReconcilerSimpleReconcile() {
	ctx := context.Background()
	resolver := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "www.google.com.br",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "www.google.com.br",
		},
	}

	reconciler := &ACLDNSEntryReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(resolver).Build(),
		Scheme:   scheme.Scheme,
		Resolver: &fakeResolver{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name: "www.google.com.br",
		},
	})
	suite.Require().NoError(err)

	existingResolver := &v1alpha1.ACLDNSEntry{}
	err = reconciler.Client.Get(ctx, client.ObjectKeyFromObject(resolver), existingResolver)
	suite.Require().NoError(err)

	suite.Require().Len(existingResolver.Status.IPs, 2)
	suite.Assert().Equal("8.8.8.8", existingResolver.Status.IPs[0].Address)
	suite.Assert().Equal("8.8.4.4", existingResolver.Status.IPs[1].Address)
}
