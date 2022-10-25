package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/tsuru/acl-operator/api/scheme"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	"github.com/tsuru/acl-operator/clients/tsuruapi"
	"github.com/tsuru/tsuru/app"
	appTypes "github.com/tsuru/tsuru/types/app"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func (suite *ControllerSuite) TestACLReconcilerSimpleReconcile() {
	ctx := context.Background()
	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Name:      "myapp",
			Namespace: "default",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "myapp",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					ExternalIP: &v1alpha1.ACLSpecExternalIP{
						IP: "100.100.100.100/32",
						Ports: v1alpha1.ACLSpecProtoPorts{
							{
								Protocol: "TCP",
								Number:   80,
							},
						},
					},
				},
			},
		},
	}

	reconciler := &ACLReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl).Build(),
		Scheme:   scheme.Scheme,
		Resolver: &fakeResolver{},
		TsuruAPI: &fakeTsuruAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "myapp",
			Namespace: "default",
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Client.Get(ctx, client.ObjectKeyFromObject(acl), existingACL)
	suite.Require().NoError(err)
	suite.Assert().True(existingACL.Status.Ready)

	existingNP := &netv1.NetworkPolicy{}
	err = reconciler.Client.Get(ctx, client.ObjectKey{
		Namespace: existingACL.Namespace,
		Name:      existingACL.Status.NetworkPolicy,
	}, existingNP)
	suite.Require().NoError(err)
	suite.Assert().Equal(map[string]string{
		"tsuru.io/app-name": "myapp",
	}, existingNP.Spec.PodSelector.MatchLabels)
}

func (suite *ControllerSuite) TestACLReconcilerDestinationAppReconcile() {
	ctx := context.Background()
	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Name:      "myapp",
			Namespace: "default",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "myapp",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					TsuruApp: "my-other-app",
				},
			},
		},
	}

	dnsEntry1 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "myapp.io",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "myapp.io",
		},
		Status: v1alpha1.ACLDNSEntryStatus{
			Ready: true,
			IPs: []v1alpha1.ACLDNSEntryStatusIP{
				{
					Address:    "1.1.1.1",
					ValidUntil: time.Now().Format(time.RFC3339),
				},
			},
		},
	}
	dnsEntry2 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "http.myapp.io",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "http.myapp.io",
		},
		Status: v1alpha1.ACLDNSEntryStatus{
			Ready: true,
			IPs: []v1alpha1.ACLDNSEntryStatusIP{
				{
					Address:    "2.2.2.2",
					ValidUntil: time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	tsuruAppAddress := &v1alpha1.TsuruAppAddress{
		ObjectMeta: v1.ObjectMeta{
			Name: "my-other-app",
		},
		Spec: v1alpha1.TsuruAppAddressSpec{
			Name: "my-other-app",
		},
		Status: v1alpha1.ResourceAddressStatus{
			Ready: true,
			IPs: []string{
				"1.1.1.1",
				"2.2.2.2",
			},
		},
	}

	reconciler := &ACLReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl, dnsEntry1, dnsEntry2, tsuruAppAddress).Build(),
		Scheme:   scheme.Scheme,
		Resolver: &fakeResolver{},
		TsuruAPI: &fakeTsuruAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "myapp",
			Namespace: "default",
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Client.Get(ctx, client.ObjectKeyFromObject(acl), existingACL)
	suite.Require().NoError(err)
	suite.Assert().True(existingACL.Status.Ready)

	existingNP := &netv1.NetworkPolicy{}
	err = reconciler.Client.Get(ctx, client.ObjectKey{
		Namespace: existingACL.Namespace,
		Name:      existingACL.Status.NetworkPolicy,
	}, existingNP)
	suite.Require().NoError(err)
	suite.Assert().Equal(map[string]string{
		"tsuru.io/app-name": "myapp",
	}, existingNP.Spec.PodSelector.MatchLabels)
	suite.Assert().Len(existingNP.Spec.Egress, 3)

	suite.Assert().Len(existingNP.Spec.Egress[0].To, 1)
	suite.Assert().Len(existingNP.Spec.Egress[1].To, 1)
	suite.Assert().Len(existingNP.Spec.Egress[2].To, 1)

	suite.Assert().Equal(netv1.NetworkPolicyPeer{
		PodSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"tsuru.io/app-name": "my-other-app",
			},
		},
	}, existingNP.Spec.Egress[0].To[0])
	suite.Assert().Equal(netv1.NetworkPolicyPeer{
		IPBlock: &netv1.IPBlock{
			CIDR: "1.1.1.1/32",
		},
	}, existingNP.Spec.Egress[1].To[0])

	suite.Assert().Equal(netv1.NetworkPolicyPeer{
		IPBlock: &netv1.IPBlock{
			CIDR: "2.2.2.2/32",
		},
	}, existingNP.Spec.Egress[2].To[0])
}

func (suite *ControllerSuite) TestACLReconcilerDestinationRPaaSReconcile() {
	ctx := context.Background()
	acl := &v1alpha1.ACL{
		ObjectMeta: v1.ObjectMeta{
			Name:      "myapp",
			Namespace: "default",
		},
		Spec: v1alpha1.ACLSpec{
			Source: v1alpha1.ACLSpecSource{
				TsuruApp: "myapp",
			},
			Destinations: []v1alpha1.ACLSpecDestination{
				{
					RpaasInstance: &v1alpha1.ACLSpecRpaasInstance{
						ServiceName: "rpaasv2",
						Instance:    "my-instance",
					},
				},
			},
		},
	}

	dnsEntry1 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "myapp.io",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "myapp.io",
		},
		Status: v1alpha1.ACLDNSEntryStatus{
			Ready: true,
			IPs: []v1alpha1.ACLDNSEntryStatusIP{
				{
					Address:    "1.1.1.1",
					ValidUntil: time.Now().Format(time.RFC3339),
				},
			},
		},
	}
	dnsEntry2 := &v1alpha1.ACLDNSEntry{
		ObjectMeta: v1.ObjectMeta{
			Name: "http.myapp.io",
		},
		Spec: v1alpha1.ACLDNSEntrySpec{
			Host: "http.myapp.io",
		},
		Status: v1alpha1.ACLDNSEntryStatus{
			Ready: true,
			IPs: []v1alpha1.ACLDNSEntryStatusIP{
				{
					Address:    "2.2.2.2",
					ValidUntil: time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	reconciler := &ACLReconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl, dnsEntry1, dnsEntry2).Build(),
		Scheme:   scheme.Scheme,
		Resolver: &fakeResolver{},
		TsuruAPI: &fakeTsuruAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "myapp",
			Namespace: "default",
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Client.Get(ctx, client.ObjectKeyFromObject(acl), existingACL)
	suite.Require().NoError(err)
	suite.Assert().True(existingACL.Status.Ready)
	suite.Assert().Equal("", existingACL.Status.Reason)
	suite.Assert().Len(existingACL.Status.WarningErrors, 0)

	existingNP := &netv1.NetworkPolicy{}
	err = reconciler.Client.Get(ctx, client.ObjectKey{
		Namespace: existingACL.Namespace,
		Name:      existingACL.Status.NetworkPolicy,
	}, existingNP)
	suite.Require().NoError(err)
	suite.Assert().Equal(map[string]string{
		"tsuru.io/app-name": "myapp",
	}, existingNP.Spec.PodSelector.MatchLabels)
	suite.Require().Len(existingNP.Spec.Egress, 2)

	suite.Assert().Len(existingNP.Spec.Egress[0].To, 1)
	suite.Assert().Len(existingNP.Spec.Egress[1].To, 1)

	suite.Assert().Equal(netv1.NetworkPolicyPeer{
		PodSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"rpaas.extensions.tsuru.io/instance-name": "my-instance",
				"rpaas.extensions.tsuru.io/service-name":  "rpaasv2",
			},
		},
	}, existingNP.Spec.Egress[0].To[0])
	suite.Assert().Equal(netv1.NetworkPolicyPeer{
		IPBlock: &netv1.IPBlock{
			CIDR: "3.3.3.3/32",
		},
	}, existingNP.Spec.Egress[1].To[0])

}

type fakeTsuruAPI struct {
}

func (f *fakeTsuruAPI) AppInfo(ctx context.Context, appName string) (*app.App, error) {
	if appName == "my-other-app" {
		return &app.App{
			Name: appName,
			Pool: "my-pool",
			Routers: []appTypes.AppRouter{
				{
					Name: "https-router",
					Addresses: []string{
						"https://myapp.io",
					},
				},
				{
					Name: "http-router",
					Addresses: []string{
						"http.myapp.io",
					},
				},
			},
		}, nil
	}

	return nil, errors.New("no app found")
}

func (f *fakeTsuruAPI) ServiceInstanceInfo(ctx context.Context, service, instance string) (*tsuruapi.ServiceInstanceInfo, error) {
	if service == "rpaasv2" && instance == "my-instance" {
		return &tsuruapi.ServiceInstanceInfo{
			CustomInfo: map[string]interface{}{
				"Address": "3.3.3.3",
			},
		}, nil
	}

	return nil, errors.New("not implemented yet")
}
