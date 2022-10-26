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
	"reflect"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	extensionstsuruiov1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	"github.com/tsuru/acl-operator/clients/tsuruapi"
)

// RpaasInstanceAddressReconciler reconciles a RpaasInstanceAddress object
type RpaasInstanceAddressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Resolver ACLDNSResolver
	TsuruAPI tsuruapi.Client
}

//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=rpaasinstanceaddresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=rpaasinstanceaddresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=rpaasinstanceaddresses/finalizers,verbs=update

func (r *RpaasInstanceAddressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	rpaasInstanceAddress := &v1alpha1.RpaasInstanceAddress{}
	err := r.Client.Get(ctx, req.NamespacedName, rpaasInstanceAddress)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		l.Error(err, "could not get RpaasInstanceAddress object")
		return ctrl.Result{}, err
	}

	serviceInfo, err := r.TsuruAPI.ServiceInstanceInfo(ctx, rpaasInstanceAddress.Spec.ServiceName, rpaasInstanceAddress.Spec.Instance)

	if err != nil {
		rpaasInstanceAddress.Status.Ready = false
		rpaasInstanceAddress.Status.Reason = err.Error()

		err = r.Client.Status().Update(ctx, rpaasInstanceAddress)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: requeueAfter,
		}, nil
	}

	if serviceInfo == nil {
		rpaasInstanceAddress.Status.Ready = false
		rpaasInstanceAddress.Status.Reason = "Service instance not found"

		err = r.Client.Status().Update(ctx, rpaasInstanceAddress)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	address := ""
	if serviceInfo.CustomInfo != nil && serviceInfo.CustomInfo["Address"] != nil {
		address, _ = serviceInfo.CustomInfo["Address"].(string)
	}

	if address == "" {
		return ctrl.Result{}, nil
	}

	var resolvedIPs []string
	if isIPRange(address) {
		resolvedIPs = []string{address}
	} else {
		// TODO: implement resolve of address
		return ctrl.Result{}, nil
	}

	if !rpaasInstanceAddress.Status.Ready || !reflect.DeepEqual(resolvedIPs, rpaasInstanceAddress.Status.IPs) {
		rpaasInstanceAddress.Status.Ready = true
		rpaasInstanceAddress.Status.IPs = resolvedIPs
		rpaasInstanceAddress.Status.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

		err = r.Client.Status().Update(ctx, rpaasInstanceAddress)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RpaasInstanceAddressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extensionstsuruiov1alpha1.RpaasInstanceAddress{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2, RecoverPanic: true}).
		Complete(r)
}
