package controllers

import (
	"context"

	"github.com/tsuru/acl-operator/api/scheme"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func (suite *ControllerSuite) TestTsuruCronJobReconcilerSimpleReconcile() {
	ctx := context.Background()
	job := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myjob",
			Namespace: "default",
			Labels: map[string]string{
				tsuruJobLabel: "myjob",
			},
		},
	}

	reconciler := &TsuruCronJobReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(job).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "myjob",
			Namespace: "default",
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      tsuruJobACLPrefix + job.Name,
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

func (suite *ControllerSuite) TestTsuruCronJobReconcilerReconcileJobWithNoRules() {
	ctx := context.Background()
	job := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myjob-no-rules",
			Namespace: "default",
			Labels: map[string]string{
				tsuruJobLabel: "myjob",
			},
		},
	}
	reconciler := &TsuruCronJobReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(job).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "myjob-no-rules",
			Namespace: "default",
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      tsuruJobACLPrefix + job.Name,
	}, existingACL)
	suite.Require().Error(err)
	suite.Require().True(k8sErrors.IsNotFound(err))
}

func (suite *ControllerSuite) TestTsuruCronJobReconcilerReconcileExistingJobWithNoRules() {
	ctx := context.Background()
	job := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myjob-no-rules",
			Namespace: "default",
			Labels: map[string]string{
				tsuruJobLabel: "myjob-no-rules",
			},
		},
	}

	acl := &v1alpha1.ACL{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: job.Namespace,
			Name:      tsuruJobACLPrefix + job.Name,
		},
		Spec: v1alpha1.ACLSpec{
			Destinations: []v1alpha1.ACLSpecDestination{},
		},
	}

	reconciler := &TsuruCronJobReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(job, acl).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: job.Namespace,
			Name:      job.Name,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      tsuruJobACLPrefix + job.Name,
	}, existingACL)
	suite.Require().Error(err)
	suite.Require().True(k8sErrors.IsNotFound(err))
}

func (suite *ControllerSuite) TestTsuruCronJobReconcilerReconcileExistingJob() {
	ctx := context.Background()
	job := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myjob",
			Namespace: "default",
			Labels: map[string]string{
				tsuruJobLabel: "myjob",
			},
		},
	}

	acl := &v1alpha1.ACL{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: job.Namespace,
			Name:      tsuruJobACLPrefix + job.Name,
		},
		Spec: v1alpha1.ACLSpec{
			Destinations: []v1alpha1.ACLSpecDestination{},
		},
	}

	reconciler := &TsuruCronJobReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(job, acl).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: job.Namespace,
			Name:      job.Name,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      tsuruJobACLPrefix + job.Name,
	}, existingACL)
	suite.Require().NoError(err)
	suite.Require().Len(existingACL.Spec.Destinations, 5)
}

func (suite *ControllerSuite) TestTsuruJobReconcilerReconcileJobWithErrors() {
	ctx := context.Background()
	job := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myjob-with-errors",
			Namespace: "default",
			Labels: map[string]string{
				tsuruJobLabel: "myjob-with-errors",
			},
		},
	}

	reconciler := &TsuruCronJobReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(job).Build(),
		Scheme: scheme.Scheme,
		ACLAPI: &fakeACLAPI{},
	}
	_, err := reconciler.Reconcile(ctx, controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      job.Name,
			Namespace: job.Namespace,
		},
	})
	suite.Require().NoError(err)

	existingACL := &v1alpha1.ACL{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      tsuruJobACLPrefix + job.Name,
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
