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
			Binds: []rpaasv1alpha1.Bind{
				{
					Name: "internal-app",
					Host: "internal-app.namespace.svc.cluster.local:8888",
				},
				{
					Name: "external-app",
					Host: "external-app.io",
				},
				{
					Name: "external-app-https",
					Host: "https://external-app.io",
				},
				{
					Name: "external-app-https-port",
					Host: "https://external-app.io:8043",
				},
				{
					Name: "external-app-http-port",
					Host: "external-app-http.io:8080",
				},
			},
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

	suite.Require().Len(existingACL.Spec.Destinations, 12)
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

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		TsuruApp: "internal-app",
	}, existingACL.Spec.Destinations[3])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		TsuruApp: "external-app",
	}, existingACL.Spec.Destinations[4])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "external-app.io",
			Ports: []v1alpha1.ProtoPort{
				{
					Protocol: "tcp",
					Number:   80,
				},
			},
		},
	}, existingACL.Spec.Destinations[5])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		TsuruApp: "external-app-https",
	}, existingACL.Spec.Destinations[6])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "external-app.io",
			Ports: []v1alpha1.ProtoPort{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[7])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		TsuruApp: "external-app-https-port",
	}, existingACL.Spec.Destinations[8])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "external-app.io",
			Ports: []v1alpha1.ProtoPort{
				{
					Protocol: "tcp",
					Number:   8043,
				},
			},
		},
	}, existingACL.Spec.Destinations[9])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		TsuruApp: "external-app-http-port",
	}, existingACL.Spec.Destinations[10])

	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "external-app-http.io",
			Ports: []v1alpha1.ProtoPort{
				{
					Protocol: "tcp",
					Number:   8080,
				},
			},
		},
	}, existingACL.Spec.Destinations[11])
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
