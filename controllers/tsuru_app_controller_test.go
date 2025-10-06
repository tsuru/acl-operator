package controllers

import (
	"context"

	"github.com/tsuru/acl-operator/api/scheme"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	"github.com/tsuru/acl-operator/clients/aclapi"
	tsuruv1 "github.com/tsuru/tsuru/provision/kubernetes/pkg/apis/tsuru/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type fakeACLAPI struct{}

func (f *fakeACLAPI) AppRules(ctx context.Context, appName string) ([]aclapi.Rule, error) {
	return f.mockRules(ctx, appName)
}

func (f *fakeACLAPI) JobRules(ctx context.Context, jobName string) ([]aclapi.Rule, error) {
	return f.mockRules(ctx, jobName)
}

func (f *fakeACLAPI) mockRules(_ context.Context, resourceName string) ([]aclapi.Rule, error) {
	switch resourceName {
	case "myapp", "myjob":
		return []aclapi.Rule{
			{
				Destination: aclapi.RuleType{
					ExternalDNS: &aclapi.ExternalDNSRule{
						Name: "www.facebook.com",
						Ports: aclapi.ProtoPorts{
							{
								Protocol: "tcp",
								Port:     80,
							},
						},
					},
				},
			},
			{
				Destination: aclapi.RuleType{
					ExternalIP: &aclapi.ExternalIPRule{
						IP: "10.1.1.1/32",
						Ports: aclapi.ProtoPorts{
							{
								Protocol: "tcp",
								Port:     443,
							},
						},
					},
				},
			},
			{
				Destination: aclapi.RuleType{
					TsuruApp: &aclapi.TsuruAppRule{
						AppName: "my-other-app",
					},
				},
			},
			{
				Destination: aclapi.RuleType{
					TsuruApp: &aclapi.TsuruAppRule{
						PoolName: "my-other-pool",
					},
				},
			},
			{
				Destination: aclapi.RuleType{
					RpaasInstance: &aclapi.RpaasInstanceRule{
						ServiceName: "my-service",
						Instance:    "my-instance",
					},
				},
			},
		}, nil
	case "myapp-no-rules", "myjob-no-rules":
		return []aclapi.Rule{}, nil
	case "myapp-with-errors", "myjob-with-errors":
		return []aclapi.Rule{
			{
				Destination: aclapi.RuleType{
					ExternalDNS: &aclapi.ExternalDNSRule{
						Name:             "www.facebook.com",
						SyncWholeNetwork: true,
					},
				},
			},
			{
				Destination: aclapi.RuleType{
					ExternalIP: &aclapi.ExternalIPRule{
						IP:               "10.1.1.1/32",
						SyncWholeNetwork: true,
					},
				},
			},
			{
				Destination: aclapi.RuleType{
					KubernetesService: &aclapi.KubernetesServiceRule{
						Namespace:   "service",
						ServiceName: "blah",
					},
				},
			},
		}, nil
	}
	return nil, nil
}

func (suite *ControllerSuite) TestTsuruAppReconcilerSimpleReconcile() {
	ctx := context.Background()
	app := &tsuruv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp",
			Namespace: "default",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "tsuru-mypool",
		},
	}

	reconciler := &TsuruAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(app).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "myapp",
			Namespace: "default",
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: app.Spec.NamespaceName,
		Name:      app.Name,
	}, existingACL)
	suite.Require().NoError(err)

	suite.Require().Len(existingACL.Spec.Destinations, 5)
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "www.facebook.com",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   80,
				},
			},
		},
	}, existingACL.Spec.Destinations[0])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalIP: &v1alpha1.ACLSpecExternalIP{
			IP: "10.1.1.1/32",
			Ports: v1alpha1.ACLSpecProtoPorts{
				{
					Protocol: "tcp",
					Number:   443,
				},
			},
		},
	}, existingACL.Spec.Destinations[1])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		TsuruApp: "my-other-app",
	}, existingACL.Spec.Destinations[2])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		TsuruAppPool: "my-other-pool",
	}, existingACL.Spec.Destinations[3])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		RpaasInstance: &v1alpha1.ACLSpecRpaasInstance{
			ServiceName: "my-service",
			Instance:    "my-instance",
		},
	}, existingACL.Spec.Destinations[4])
}

func (suite *ControllerSuite) TestTsuruAppReconcilerReconcileAppWithNoRules() {
	ctx := context.Background()
	app := &tsuruv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-no-rules",
			Namespace: "default",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "tsuru-mypool",
		},
	}

	reconciler := &TsuruAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(app).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: app.Spec.NamespaceName,
		Name:      app.Name,
	}, existingACL)
	suite.Require().Error(err)
	suite.Require().True(k8sErrors.IsNotFound(err))
}

func (suite *ControllerSuite) TestTsuruAppReconcilerReconcileExistingAppWithNoRules() {
	ctx := context.Background()
	app := &tsuruv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-no-rules",
			Namespace: "default",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "tsuru-mypool",
		},
	}

	acl := &v1alpha1.ACL{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: app.Spec.NamespaceName,
			Name:      app.Name,
		},
		Spec: v1alpha1.ACLSpec{
			Destinations: []v1alpha1.ACLSpecDestination{},
		},
	}

	reconciler := &TsuruAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(app, acl).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: app.Spec.NamespaceName,
		Name:      app.Name,
	}, existingACL)
	suite.Require().Error(err)
	suite.Require().True(k8sErrors.IsNotFound(err))
}

func (suite *ControllerSuite) TestTsuruAppReconcilerReconcileExistingApp() {
	ctx := context.Background()
	app := &tsuruv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp",
			Namespace: "default",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "tsuru-mypool",
		},
	}

	acl := &v1alpha1.ACL{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: app.Spec.NamespaceName,
			Name:      app.Name,
		},
		Spec: v1alpha1.ACLSpec{
			Destinations: []v1alpha1.ACLSpecDestination{},
		},
	}

	reconciler := &TsuruAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(app, acl).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: app.Spec.NamespaceName,
		Name:      app.Name,
	}, existingACL)
	suite.Require().NoError(err)
	suite.Require().Len(existingACL.Spec.Destinations, 5)
}

func (suite *ControllerSuite) TestTsuruAppReconcilerReconcileAppWithErrors() {
	ctx := context.Background()
	app := &tsuruv1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-with-errors",
			Namespace: "default",
		},
		Spec: tsuruv1.AppSpec{
			NamespaceName: "tsuru-mypool",
		},
	}

	reconciler := &TsuruAppReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(app).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: app.Spec.NamespaceName,
		Name:      app.Name,
	}, existingACL)
	suite.Require().NoError(err)
	suite.Assert().Len(existingACL.Spec.Destinations, 2)
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
			Name: "www.facebook.com",
		},
	}, existingACL.Spec.Destinations[0])
	suite.Assert().Equal(v1alpha1.ACLSpecDestination{
		ExternalIP: &v1alpha1.ACLSpecExternalIP{
			IP: "10.1.1.1/32",
		},
	}, existingACL.Spec.Destinations[1])

	suite.Assert().Len(existingACL.Status.WarningErrors, 3)
}
