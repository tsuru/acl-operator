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
	"errors"
	"net"
	"reflect"
	"sort"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tsuru/acl-operator/api/v1alpha1"
	extensionstsuruiov1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	"github.com/tsuru/acl-operator/clients/tsuruapi"
	tsuruNet "github.com/tsuru/tsuru/net"
)

var errAppNotFound = errors.New("App not found")

// TsuruAppAddressReconciler reconciles a TsuruAppAddress object
type TsuruAppAddressReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Resolver ACLDNSResolver
	TsuruAPI tsuruapi.Client
}

//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=tsuruappaddresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=tsuruappaddresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=tsuruappaddresses/finalizers,verbs=update

func (r *TsuruAppAddressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	appAddress := &v1alpha1.TsuruAppAddress{}
	err := r.Client.Get(ctx, req.NamespacedName, appAddress)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		l.Error(err, "could not get TsuruAppAddress object")
		return ctrl.Result{}, err
	}

	oldStatus := appAddress.Status.DeepCopy()
	err = r.FillStatus(ctx, appAddress)
	if err != nil {
		appAddress.Status.Ready = false
		appAddress.Status.Reason = err.Error()
	}

	if oldStatus.Ready != appAddress.Status.Ready || !reflect.DeepEqual(oldStatus.IPs, appAddress.Status.IPs) {
		err = r.Client.Status().Update(ctx, appAddress)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TsuruAppAddressReconciler) FillStatus(ctx context.Context, appAddress *v1alpha1.TsuruAppAddress) error {
	appInfo, err := r.TsuruAPI.AppInfo(ctx, appAddress.Spec.Name)
	if err != nil {
		return err
	}

	if appInfo == nil {
		return errAppNotFound
	}

	addrs := make([]string, 0, len(appInfo.Routers))
	for _, r := range appInfo.Routers {
		if len(r.Addresses) > 0 {
			for _, addr := range r.Addresses {
				addrs = append(addrs, tsuruNet.URLToHost(addr))
			}
		} else {
			addrs = append(addrs, tsuruNet.URLToHost(r.Address))
		}
	}

	foundIPs := map[string]bool{}
	for _, addr := range addrs {
		ipAddrs, err := r.resolveAddress(ctx, addr)
		if err != nil {
			// TODO: set feedback on app
			continue
		}

		for _, ipAddr := range ipAddrs {
			foundIPs[ipAddr.IP.String()] = true
		}
	}

	resolvedIPs := []string{}
	for ip := range foundIPs {
		resolvedIPs = append(resolvedIPs, ip)
	}
	sort.Strings(resolvedIPs)

	if !appAddress.Status.Ready || !reflect.DeepEqual(resolvedIPs, appAddress.Status.IPs) {
		appAddress.Status.Ready = true
		appAddress.Status.Reason = ""
		appAddress.Status.IPs = resolvedIPs
		appAddress.Status.Pool = appInfo.Pool
		appAddress.Status.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	return nil
}

func (r *TsuruAppAddressReconciler) resolveAddress(ctx context.Context, addr string) ([]net.IPAddr, error) {
	timoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return r.Resolver.LookupIPAddr(timoutCtx, addr)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TsuruAppAddressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extensionstsuruiov1alpha1.TsuruAppAddress{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2, RecoverPanic: true}).
		Complete(r)
}
