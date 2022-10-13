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
	"fmt"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tsuruv1 "github.com/tsuru/tsuru/provision/kubernetes/pkg/apis/tsuru/v1"

	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	aclapi "github.com/tsuru/acl-operator/clients/aclapi"
)

// TsuruAppReconciler reconciles a Tsuru App object
type TsuruAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ACLAPI aclapi.Client
}

func (r *TsuruAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	app := &tsuruv1.App{}

	err := r.Client.Get(ctx, req.NamespacedName, app)
	if err != nil {
		l.Error(err, "could not get Tsuru App object")
		return ctrl.Result{}, err
	}

	rules, err := r.ACLAPI.AppRules(ctx, app.Name)
	if err != nil {
		l.Error(err, "could not get Tsuru App Rules from ACLAPI")
		return ctrl.Result{}, err
	}

	destinations, errs := convertACLAPIRulesToOperatorRules(rules)
	if len(errs) > 0 {

	}

	acl := &v1alpha1.ACL{}
	err = r.Client.Get(ctx, client.ObjectKey{
		Name:      app.Name,
		Namespace: app.Spec.NamespaceName,
	}, acl)

	if k8sErrors.IsNotFound(err) {
		if len(destinations) == 0 {
			return ctrl.Result{}, nil
		}

		err = r.Client.Create(ctx, &v1alpha1.ACL{
			ObjectMeta: metav1.ObjectMeta{
				Name:      app.Name,
				Namespace: app.Spec.NamespaceName,
			},
			Spec: v1alpha1.ACLSpec{
				Source: v1alpha1.ACLSpecSource{
					TsuruApp: app.Name,
				},
				Destinations: destinations,
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

	acl.Spec.Source = v1alpha1.ACLSpecSource{
		TsuruApp: app.Name,
	}
	acl.Spec.Destinations = destinations

	err = r.Client.Update(ctx, acl)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: requeueAfter,
	}, nil
}

func convertACLAPIRulesToOperatorRules(rules []aclapi.Rule) ([]v1alpha1.ACLSpecDestination, []error) {
	result := []v1alpha1.ACLSpecDestination{}
	errors := []error{}

	for _, rule := range rules {
		if rule.Removed {
			continue
		}

		if rule.Destination.TsuruApp != nil {
			if rule.Destination.TsuruApp.AppName != "" {
				result = append(result, v1alpha1.ACLSpecDestination{
					TsuruApp: rule.Destination.TsuruApp.AppName,
				})
			} else if rule.Destination.TsuruApp.PoolName != "" {
				result = append(result, v1alpha1.ACLSpecDestination{
					TsuruAppPool: rule.Destination.TsuruApp.PoolName,
				})
			}
		} else if rule.Destination.ExternalDNS != nil {
			externalDNS, err := convertExternalDNSDestination(rule.Destination.ExternalDNS)
			if err != nil {
				errors = append(errors, err)
			}
			result = append(result, v1alpha1.ACLSpecDestination{
				ExternalDNS: externalDNS,
			})
		} else if rule.Destination.ExternalIP != nil {
			externalIP, err := convertExternalIPDestination(rule.Destination.ExternalIP)
			if err != nil {
				errors = append(errors, err)
			}
			result = append(result, v1alpha1.ACLSpecDestination{
				ExternalIP: externalIP,
			})
		} else if rule.Destination.RpaasInstance != nil {
			rpaasInstance, err := convertRpaasInstanceDestination(rule.Destination.RpaasInstance)
			if err != nil {
				errors = append(errors, err)
			}
			result = append(result, v1alpha1.ACLSpecDestination{
				RpaasInstance: rpaasInstance,
			})
		} else if rule.Destination.KubernetesService != nil {
			err := fmt.Errorf("Kubernetes service is not supported yet")
			errors = append(errors, err)
		}
	}

	return result, errors
}

func convertExternalDNSDestination(rule *aclapi.ExternalDNSRule) (*v1alpha1.ACLSpecExternalDNS, error) {
	result := &v1alpha1.ACLSpecExternalDNS{
		Name:  rule.Name,
		Ports: convertPorts(rule.Ports),
	}

	if rule.SyncWholeNetwork {
		return nil, fmt.Errorf("SyncWholeNetwork is not supported for %q", rule.Name)
	}

	return result, nil
}

func convertExternalIPDestination(rule *aclapi.ExternalIPRule) (*v1alpha1.ACLSpecExternalIP, error) {
	result := &v1alpha1.ACLSpecExternalIP{
		IP:    rule.IP,
		Ports: convertPorts(rule.Ports),
	}

	if rule.SyncWholeNetwork {
		return nil, fmt.Errorf("SyncWholeNetwork is not supported for %q", rule.IP)
	}

	return result, nil
}

func convertRpaasInstanceDestination(rule *aclapi.RpaasInstanceRule) (*v1alpha1.ACLSpecRpaasInstance, error) {
	result := &v1alpha1.ACLSpecRpaasInstance{
		ServiceName: rule.ServiceName,
		Instance:    rule.Instance,
	}
	return result, nil
}

func convertPorts(ports aclapi.ProtoPorts) v1alpha1.ACLSpecProtoPorts {
	result := v1alpha1.ACLSpecProtoPorts{}

	for _, port := range ports {
		result = append(result, v1alpha1.ProtoPort{
			Protocol: port.Protocol,
			Number:   port.Port,
		})
	}

	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *TsuruAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tsuruv1.App{}).
		Complete(r)
}
