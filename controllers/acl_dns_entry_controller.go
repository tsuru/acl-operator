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
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

func (r *ACLDNSEntryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	dnsEntry := &v1alpha1.ACLDNSEntry{}

	err := r.Client.Get(ctx, req.NamespacedName, dnsEntry)
	if k8sErrors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		l.Error(err, "could not get ACLDNSEntry object")
		return ctrl.Result{}, err
	}

	existingStatus := dnsEntry.Status.DeepCopy()

	err = r.FillStatus(ctx, dnsEntry)

	if err != nil {
		l.Error(err, "could not resolve address", "host", dnsEntry.Spec.Host)

		dnsEntry.Status.Ready = false
		dnsEntry.Status.Reason = err.Error()

		statusErr := r.Client.Status().Update(ctx, dnsEntry)
		if statusErr != nil {
			l.Error(statusErr, "could not update status for ACLDNSEntry object")
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Minute * 10,
		}, nil
	}

	if !reflect.DeepEqual(existingStatus, dnsEntry.Status) {
		err = r.Client.Status().Update(ctx, dnsEntry)
		if err != nil {
			l.Error(err, "could not update status for ACLDNSEntry object")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ACLDNSEntryReconciler) FillStatus(ctx context.Context, dnsEntry *v1alpha1.ACLDNSEntry) error {
	timoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ipAddrs, err := r.Resolver.LookupIPAddr(timoutCtx, dnsEntry.Spec.Host)

	if err != nil {
		return err
	}

	now := time.Now().UTC()
	validUntil := now.Add(7 * 24 * time.Hour)

	missingIpAddrs := []net.IPAddr{}
statusLoop:
	for _, foundIP := range ipAddrs {
		for i, existingIP := range dnsEntry.Status.IPs {
			if existingIP.Address == foundIP.IP.String() {
				dnsEntry.Status.IPs[i].ValidUntil = validUntil.Format(dayFormat)
				continue statusLoop
			}
		}

		missingIpAddrs = append(missingIpAddrs, foundIP)
	}

	for _, foundIP := range missingIpAddrs {
		dnsEntry.Status.IPs = append(dnsEntry.Status.IPs, extensionstsuruiov1alpha1.ACLDNSEntryStatusIP{
			Address:    foundIP.IP.String(),
			ValidUntil: validUntil.Format(dayFormat),
		})
	}

	sort.Slice(dnsEntry.Status.IPs, func(i, j int) bool {
		return dnsEntry.Status.IPs[i].Address < dnsEntry.Status.IPs[j].Address
	})

	n := 0
	for _, ip := range dnsEntry.Status.IPs {
		t, _ := time.Parse(dayFormat, ip.ValidUntil)

		if !now.After(t) && !t.IsZero() {
			dnsEntry.Status.IPs[n] = ip
			n++
		}
	}
	dnsEntry.Status.IPs = dnsEntry.Status.IPs[:n]
	dnsEntry.Status.Ready = true
	dnsEntry.Status.Reason = ""

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ACLDNSEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&extensionstsuruiov1alpha1.ACLDNSEntry{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 4, RecoverPanic: true}).
		Complete(r)
}
