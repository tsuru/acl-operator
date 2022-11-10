package controllers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tsuru/acl-operator/api/scheme"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	"github.com/tsuru/acl-operator/clients/tsuruapi"
	"github.com/tsuru/tsuru/app"
	appTypes "github.com/tsuru/tsuru/types/app"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
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
					RuleID: "external-ip-1",
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
				{
					RuleID: "external-ip-2",
					ExternalIP: &v1alpha1.ACLSpecExternalIP{
						IP: "1.1.1.1/32",
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
	tcp := corev1.ProtocolTCP

	err = reconciler.Client.Get(ctx, client.ObjectKeyFromObject(acl), existingACL)
	suite.Require().NoError(err)
	suite.Assert().True(existingACL.Status.Ready)
	suite.Assert().Equal("", existingACL.Status.Reason)
	suite.Assert().Equal([]v1alpha1.ACLStatusStale{
		{
			RuleID: "external-ip-1",
			Rules: []netv1.NetworkPolicyEgressRule{
				{
					Ports: []netv1.NetworkPolicyPort{
						{
							Port: &intstr.IntOrString{
								IntVal: 80,
							},
							Protocol: &tcp,
						},
					},
					To: []netv1.NetworkPolicyPeer{
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "100.100.100.100/32",
							},
						},
					},
				},
			},
		},
		{
			RuleID: "external-ip-2",
			Rules: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "1.1.1.1/32",
							},
						},
					},
				},
			},
		},
	}, existingACL.Status.Stale)
	suite.Assert().Len(existingACL.Status.RuleErrors, 0)

	existingNP := &netv1.NetworkPolicy{}
	err = reconciler.Client.Get(ctx, client.ObjectKey{
		Namespace: existingACL.Namespace,
		Name:      existingACL.Status.NetworkPolicy,
	}, existingNP)
	suite.Require().NoError(err)
	suite.Assert().Equal(map[string]string{
		"tsuru.io/app-name": "myapp",
	}, existingNP.Spec.PodSelector.MatchLabels)
	suite.Assert().Equal(netv1.NetworkPolicyEgressRule{
		Ports: []netv1.NetworkPolicyPort{
			{
				Port: &intstr.IntOrString{
					IntVal: 80,
				},
				Protocol: &tcp,
			},
		},
		To: []netv1.NetworkPolicyPeer{
			{
				IPBlock: &netv1.IPBlock{
					CIDR: "100.100.100.100/32",
				},
			},
		},
	}, existingNP.Spec.Egress[0])

	suite.Assert().Equal(netv1.NetworkPolicyEgressRule{
		To: []netv1.NetworkPolicyPeer{
			{
				IPBlock: &netv1.IPBlock{
					CIDR: "1.1.1.1/32",
				},
			},
		},
	}, existingNP.Spec.Egress[1])
}

func (suite *ControllerSuite) TestACLReconcilerStaleReconcile() {
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
					RuleID: "external-ip-2",
					ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
						Name: "timeout.com.br",
					},
				},
			},
		},
		Status: v1alpha1.ACLStatus{
			Stale: []v1alpha1.ACLStatusStale{
				{
					RuleID: "external-ip-2",
					Rules: []netv1.NetworkPolicyEgressRule{
						{
							To: []netv1.NetworkPolicyPeer{
								{
									IPBlock: &netv1.IPBlock{
										CIDR: "200.200.200.200/32",
									},
								},
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
	suite.Assert().Equal("", existingACL.Status.Reason)
	suite.Assert().Equal([]v1alpha1.ACLStatusStale{
		{
			RuleID: "external-ip-2",
			Rules: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{
						{
							IPBlock: &netv1.IPBlock{
								CIDR: "200.200.200.200/32",
							},
						},
					},
				},
			},
		},
	}, existingACL.Status.Stale)
	suite.Require().Len(existingACL.Status.RuleErrors, 1)
	suite.Assert().Equal(v1alpha1.ACLStatusRuleError{
		RuleID: "external-ip-2",
		Error:  "timeout for host",
	}, existingACL.Status.RuleErrors[0])

	existingNP := &netv1.NetworkPolicy{}
	err = reconciler.Client.Get(ctx, client.ObjectKey{
		Namespace: existingACL.Namespace,
		Name:      existingACL.Status.NetworkPolicy,
	}, existingNP)
	suite.Require().NoError(err)
	suite.Assert().Equal(map[string]string{
		"tsuru.io/app-name": "myapp",
	}, existingNP.Spec.PodSelector.MatchLabels)
	suite.Assert().Equal(netv1.NetworkPolicyEgressRule{
		To: []netv1.NetworkPolicyPeer{
			{
				IPBlock: &netv1.IPBlock{
					CIDR: "200.200.200.200/32",
				},
			},
		},
	}, existingNP.Spec.Egress[0])
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

	// 1.1.1.1 is also running on kubernetes
	svc := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-awesome-service",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Selector: map[string]string{
				"svc": "my-awesome-service",
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						IP: "1.1.1.1",
					},
				},
			},
		},
	}

	reconciler := &ACLReconciler{
		Client: fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithRuntimeObjects(acl, dnsEntry1, dnsEntry2, tsuruAppAddress, svc).
			Build(),
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
	suite.Assert().Len(existingNP.Spec.Egress[1].To, 2)
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
		PodSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"svc": "my-awesome-service",
			},
		},
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": "default",
			},
		},
	}, existingNP.Spec.Egress[1].To[1])

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

func TestValidResourceName(t *testing.T) {
	expectations := map[string]string{
		"user":         "user",
		"*.globo.com":  "globo.com-102f523825",
		".google.com":  "google.com-5d59719991",
		"10.1.1.1/10":  "10.1.1.1-10-22f870d4a0",
		"facebook.com": "facebook.com",

		strings.Repeat("testing-", 30): "testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing-testing--2744fb94f9",
	}

	for input, expectedOutput := range expectations {
		output := validResourceName(input)
		assert.Equal(t, expectedOutput, output, "input", input)

		errs := validation.IsDNS1123Subdomain(output)

		assert.Len(t, errs, 0)
	}

}
