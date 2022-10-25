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
	"sigs.k8s.io/controller-runtime/pkg/log"

	extensionstsuruiov1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	"github.com/tsuru/acl-operator/clients/tsuruapi"
)

// RpaasInstanceAdressReconciler reconciles a RpaasInstanceAdress object
type RpaasInstanceAdressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Resolver ACLDNSResolver
	TsuruAPI tsuruapi.Client
}

//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=rpaasinstanceadresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=rpaasinstanceadresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=rpaasinstanceadresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RpaasInstanceAdress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *RpaasInstanceAdressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	rpaasInstanceAddress := &v1alpha1.RpaasInstanceAdress{}
	err := r.Client.Get(ctx, req.NamespacedName, rpaasInstanceAddress)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		l.Error(err, "could not get RpaasInstanceAdress object")
		return ctrl.Result{}, err
	}

	serviceInfo, err := r.TsuruAPI.ServiceInstanceInfo(ctx, rpaasInstanceAddress.Spec.ServiceName, rpaasInstanceAddress.Spec.Instance)

	if err != nil {
		// TODO: update object status
		return ctrl.Result{}, err
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
func (r *RpaasInstanceAdressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extensionstsuruiov1alpha1.RpaasInstanceAdress{}).
		Complete(r)
}
