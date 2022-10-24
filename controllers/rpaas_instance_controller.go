/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	rpaasv1alpha1 "github.com/tsuru/rpaas-operator/api/v1alpha1"
)

// TsuruAppReconciler reconciles a Tsuru App object
type RpaasInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RpaasInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	rpaasInstance := &rpaasv1alpha1.RpaasInstance{}

	err := r.Client.Get(ctx, req.NamespacedName, rpaasInstance)
	if err != nil {
		l.Error(err, "could not get RPaaS Instance object")
		return ctrl.Result{}, err
	}

	rpaasInstanceName := rpaasInstance.Labels["rpaas.extensions.tsuru.io/instance-name"]
	rpaasServiceName := rpaasInstance.Labels["rpaas.extensions.tsuru.io/service-name"]

	destinations, errs := convertRPaasAllowedUpstreamsToOperatorRules(rpaasInstance.Spec.AllowedUpstreams, rpaasInstance.Spec.Binds)
	warningErrors := []string{}
	for _, e := range errs {
		warningErrors = append(warningErrors, e.Error())
	}

	acl := &v1alpha1.ACL{}
	err = r.Client.Get(ctx, client.ObjectKey{
		Name:      rpaasInstance.Name,
		Namespace: rpaasInstance.Namespace,
	}, acl)

	if k8sErrors.IsNotFound(err) {
		if len(rpaasInstance.Spec.AllowedUpstreams) == 0 {
			return ctrl.Result{}, nil
		}

		err = r.Client.Create(ctx, &v1alpha1.ACL{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rpaasInstance.Name,
				Namespace: rpaasInstance.Namespace,
				OwnerReferences: []v1.OwnerReference{
					*metav1.NewControllerRef(rpaasInstance, rpaasInstance.GroupVersionKind()),
				},
			},
			Spec: v1alpha1.ACLSpec{
				Source: v1alpha1.ACLSpecSource{
					RpaasInstance: &v1alpha1.ACLSpecRpaasInstance{
						ServiceName: rpaasServiceName,
						Instance:    rpaasInstanceName,
					},
				},
				Destinations: destinations,
			},
			Status: v1alpha1.ACLStatus{
				WarningErrors: warningErrors,
			},
		})

		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: requeueAfter,
		}, nil
	} else if err != nil {
		l.Error(err, "could not get ACL object")
		return ctrl.Result{}, err
	} else if len(destinations) == 0 {
		err = r.Client.Delete(ctx, acl)

		if err != nil {
			l.Error(err, "could not remove unused ACL")
		}
		return ctrl.Result{}, nil
	}

	acl.OwnerReferences = []v1.OwnerReference{
		*metav1.NewControllerRef(rpaasInstance, rpaasInstance.GroupVersionKind()),
	}

	acl.Spec.Source = v1alpha1.ACLSpecSource{
		RpaasInstance: &v1alpha1.ACLSpecRpaasInstance{
			ServiceName: rpaasServiceName,
			Instance:    rpaasInstanceName,
		},
	}
	acl.Spec.Destinations = destinations

	err = r.Client.Update(ctx, acl)
	if err != nil {
		return ctrl.Result{}, err
	}

	if len(warningErrors) > 0 || len(acl.Status.WarningErrors) > 0 {
		acl.Status.WarningErrors = warningErrors

		err := r.Client.Status().Update(ctx, acl)
		if err != nil {
			l.Error(err, "could not remove update status of ACL")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: requeueAfter,
	}, nil
}

func convertRPaasAllowedUpstreamsToOperatorRules(allowedUpstreams []rpaasv1alpha1.AllowedUpstream, binds []rpaasv1alpha1.Bind) ([]v1alpha1.ACLSpecDestination, []error) {
	result := []v1alpha1.ACLSpecDestination{}
	for _, allowedUpstream := range allowedUpstreams {
		if isIPRange(allowedUpstream.Host) {
			result = append(result, v1alpha1.ACLSpecDestination{
				ExternalIP: &v1alpha1.ACLSpecExternalIP{
					IP: allowedUpstream.Host,
					Ports: []v1alpha1.ProtoPort{
						{
							Protocol: "tcp",
							Number:   uint16(allowedUpstream.Port),
						},
					},
				},
			})
		} else {
			result = append(result, v1alpha1.ACLSpecDestination{
				ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
					Name: allowedUpstream.Host,
					Ports: []v1alpha1.ProtoPort{
						{
							Protocol: "tcp",
							Number:   uint16(allowedUpstream.Port),
						},
					},
				},
			})
		}
	}

	for _, bind := range binds {
		result = append(result, v1alpha1.ACLSpecDestination{
			TsuruApp: bind.Name,
		})

		if !isKubernetesInternal(bind.Host) {
			host, port, _ := parseHostPort(bind.Host)

			if host == "" {
				continue
			}

			var ports []v1alpha1.ProtoPort
			if port != 0 {
				ports = []v1alpha1.ProtoPort{
					{
						Protocol: "tcp",
						Number:   uint16(port),
					},
				}
			}

			result = append(result, v1alpha1.ACLSpecDestination{
				ExternalDNS: &v1alpha1.ACLSpecExternalDNS{
					Name:  host,
					Ports: ports,
				},
			})
		}
	}

	return result, nil
}

func isIPRange(name string) bool {
	var address string
	if strings.Contains(name, "/") {
		address = name
	} else {
		address = name + "/32"
	}

	_, _, err := net.ParseCIDR(address)
	return err == nil
}

func parseHostPort(addr string) (string, int, error) {
	port := 80
	host := ""
	u, err := url.Parse(addr)
	if err != nil {
		return "", 0, err
	}

	if u != nil {
		if u.Scheme == "https" {
			port = 443
		}

		if u.Host != "" {
			host = u.Host
		}
	}

	if strings.Contains(host, ":") {
		var portStr string
		host, portStr, _ = net.SplitHostPort(host)
		port, _ = strconv.Atoi(portStr)
	}

	if host == "" && strings.Contains(addr, ":") { // could be a host:port pair
		var portStr string
		host, portStr, _ = net.SplitHostPort(addr)
		port, _ = strconv.Atoi(portStr)
	}

	if host == "" {
		host = addr
	}

	return host, port, nil
}

func isKubernetesInternal(host string) bool {
	return strings.Contains(host, "svc.cluster.local")
}

// SetupWithManager sets up the controller with the Manager.
func (r *RpaasInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rpaasv1alpha1.RpaasInstance{}).
		Complete(r)
}
