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
	"reflect"
	"sort"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tsuru/acl-operator/api/v1alpha1"
	extensionstsuruiov1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
)

const dayFormat = "2006-01-02"

type ACLDNSResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

var DefaultResolver ACLDNSResolver = &net.Resolver{}

// ACLDNSEntryReconciler reconciles a ACLDNSEntry object
type ACLDNSEntryReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Resolver ACLDNSResolver
}

//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=ACLDNSEntrys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=ACLDNSEntrys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=extensions.tsuru.io,resources=ACLDNSEntrys/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ACLDNSEntry object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ACLDNSEntryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	resolver := &v1alpha1.ACLDNSEntry{}

	err := r.Client.Get(ctx, req.NamespacedName, resolver)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		l.Error(err, "could not get ACLDNSEntry object")
		return ctrl.Result{}, err
	}

	existingStatus := resolver.Status.DeepCopy()

	timoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ipAddrs, err := r.Resolver.LookupIPAddr(timoutCtx, resolver.Spec.Host)

	if err != nil {
		l.Error(err, "could not resolve address", "host", resolver.Spec.Host)

		resolver.Status.Ready = false
		resolver.Status.Reason = err.Error()

		statusErr := r.Client.Status().Update(ctx, resolver)
		if statusErr != nil {
			l.Error(statusErr, "could not update status for ACLDNSEntry object")
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}

	now := time.Now().UTC()
	validUntil := now.Add(7 * 24 * time.Hour)

	missingIpAddrs := []net.IPAddr{}
statusLoop:
	for _, foundIP := range ipAddrs {
		for i, existingIP := range resolver.Status.IPs {
			if existingIP.Address == foundIP.IP.String() {
				resolver.Status.IPs[i].ValidUntil = validUntil.Format(dayFormat)
				continue statusLoop
			}
		}

		missingIpAddrs = append(missingIpAddrs, foundIP)
	}

	for _, foundIP := range missingIpAddrs {
		resolver.Status.IPs = append(resolver.Status.IPs, extensionstsuruiov1alpha1.ACLDNSEntryStatusIP{
			Address:    foundIP.IP.String(),
			ValidUntil: validUntil.Format(dayFormat),
		})
	}

	sort.Slice(resolver.Status.IPs, func(i, j int) bool {
		return resolver.Status.IPs[i].Address < resolver.Status.IPs[j].Address
	})

	n := 0
	for _, ip := range resolver.Status.IPs {
		t, _ := time.Parse(dayFormat, ip.ValidUntil)

		if !now.After(t) && !t.IsZero() {
			resolver.Status.IPs[n] = ip
			n++
		}
	}
	resolver.Status.IPs = resolver.Status.IPs[:n]
	resolver.Status.Ready = true
	resolver.Status.Reason = ""

	if !reflect.DeepEqual(existingStatus, resolver.Status) {
		err = r.Client.Status().Update(ctx, resolver)
		if err != nil {
			l.Error(err, "could not update status for ACLDNSEntry object")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ACLDNSEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extensionstsuruiov1alpha1.ACLDNSEntry{}).
		Complete(r)
}
