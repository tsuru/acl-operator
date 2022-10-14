package controllers

import (
	"context"

	"github.com/tsuru/acl-operator/api/scheme"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	rpaasv1alpha1 "github.com/tsuru/rpaas-operator/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func (suite *ControllerSuite) TestRPaaSInstanceReconcilerSimpleReconcile() {
	ctx := context.Background()
	rpaasInstance := &rpaasv1alpha1.RpaasInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-rpaas",
			Namespace: "default",
			Labels: map[string]string{
				"rpaas.extensions.tsuru.io/instance-name": "my-rpaas",
				"rpaas.extensions.tsuru.io/service-name":  "my-service",
			},
		},
		Spec: rpaasv1alpha1.RpaasInstanceSpec{
			AllowedUpstreams: []rpaasv1alpha1.AllowedUpstream{
				{
					Host: "www.facebook.com",
					Port: 443,
				},
				{
					Host: "10.1.1.1",
					Port: 443,
				},
				{
					Host: "192.168.1.0/24",
					Port: 443,
				},
			},
		},
	}

	reconciler := &RpaasInstanceReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(rpaasInstance).Build(),
		Scheme: scheme.Scheme,
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      rpaasInstance.Name,
			Namespace: rpaasInstance.Namespace,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Client.Get(ctx, types.NamespacedName{
		Name:      rpaasInstance.Name,
		Namespace: rpaasInstance.Namespace,
	}, existingACL)
	suite.Require().NoError(err)

	suite.Require().Len(existingACL.Spec.Destinations, 3)
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "www.facebook.com",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[0])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalIP: &v1alpha1.ACLSpecExternalIP{
			IP: "10.1.1.1",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[1])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalIP: &v1alpha1.ACLSpecExternalIP{
			IP: "192.168.1.0/24",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[2])
}

func (suite *ControllerSuite) TestRPaaSInstanceReconcilerExistingObjectReconcile() {
	ctx := context.Background()
	rpaasInstance := &rpaasv1alpha1.RpaasInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-rpaas",
			Namespace: "default",
			Labels: map[string]string{
				"rpaas.extensions.tsuru.io/instance-name": "my-rpaas",
				"rpaas.extensions.tsuru.io/service-name":  "my-service",
			},
		},
		Spec: rpaasv1alpha1.RpaasInstanceSpec{
			AllowedUpstreams: []rpaasv1alpha1.AllowedUpstream{
				{
					Host: "www.facebook.com",
					Port: 443,
				},
				{
					Host: "10.1.1.1",
					Port: 443,
				},
				{
					Host: "192.168.1.0/24",
					Port: 443,
				},
			},
		},
	}

	acl := &v1alpha1.ACL{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rpaasInstance.Namespace,
			Name:      rpaasInstance.Name,
		},
		Spec: v1alpha1.ACLSpec{
			Destinations: []v1alpha1.ACLSpecDestination{},
		},
	}

	reconciler := &RpaasInstanceReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(rpaasInstance, acl).Build(),
		Scheme: scheme.Scheme,
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      rpaasInstance.Name,
			Namespace: rpaasInstance.Namespace,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Client.Get(ctx, types.NamespacedName{
		Name:      rpaasInstance.Name,
		Namespace: rpaasInstance.Namespace,
	}, existingACL)
	suite.Require().NoError(err)

	suite.Require().Len(existingACL.Spec.Destinations, 3)
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "www.facebook.com",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[0])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalIP: &v1alpha1.ACLSpecExternalIP{
			IP: "10.1.1.1",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[1])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalIP: &v1alpha1.ACLSpecExternalIP{
			IP: "192.168.1.0/24",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[2])
}
