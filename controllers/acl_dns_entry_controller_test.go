package controllers

import (
	"context"
	"errors"
	"net"

	"github.com/tsuru/acl-operator/api/scheme"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

	if host == "timeout.com.br" {
		return nil, errors.New("timeout for host")

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

	suite.Assert().True(existingResolver.Status.Ready)
	suite.Require().Len(existingResolver.Status.IPs, 2)
	suite.Assert().Equal("8.8.4.4", existingResolver.Status.IPs[0].Address)
	suite.Assert().Equal("8.8.8.8", existingResolver.Status.IPs[1].Address)
}

func (suite *ControllerSuite) TestACLDNSEntryReconcilerSimpleReconcileExisting() {
	ctx := context.Background()
	resolver := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "www.google.com.br",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "www.google.com.br",
		},
		Status: v1alpha1.ACLDNSEntryStatus{
			IPs: []v1alpha1.ACLDNSEntryStatusIP{
				{
					Address:    "1.1.1.1",
					ValidUntil: "2015-10-02",
				},
				{
					Address:    "8.8.8.8",
					ValidUntil: "2015-10-02",
				},
				{
					Address:    "9.9.9.9",
					ValidUntil: "2200-10-02", // I expect that someone of future wont judge-me
				},
			},
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

	suite.Assert().True(existingResolver.Status.Ready)
	suite.Require().Len(existingResolver.Status.IPs, 3)
	suite.Assert().Equal("8.8.4.4", existingResolver.Status.IPs[0].Address)
	suite.Assert().Equal("8.8.8.8", existingResolver.Status.IPs[1].Address)
	suite.Assert().Equal("9.9.9.9", existingResolver.Status.IPs[2].Address)
}

func (suite *ControllerSuite) TestACLDNSEntryReconcilerTimeoutReconcile() {
	ctx := context.Background()
	resolver := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "timeout.com.br",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "timeout.com.br",
		},
	}

	reconciler := &ACLDNSEntryReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(resolver).Build(),
		Scheme:   scheme.Scheme,
		Resolver: &fakeResolver{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name: "timeout.com.br",
		},
	})
	suite.Require().NoError(err)

	existingResolver := &v1alpha1.ACLDNSEntry{}
	err = reconciler.Client.Get(ctx, client.ObjectKeyFromObject(resolver), existingResolver)
	suite.Require().NoError(err)

	suite.Require().Len(existingResolver.Status.IPs, 0)
	suite.Assert().False(existingResolver.Status.Ready)
	suite.Assert().Equal("timeout for host", existingResolver.Status.Reason)
}
