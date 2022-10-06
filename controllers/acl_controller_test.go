package controllers

import (
	"context"

	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
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
		Client: fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(acl).Build(),
		Scheme: scheme.Scheme,
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
