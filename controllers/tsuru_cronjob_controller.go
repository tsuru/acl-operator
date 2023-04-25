/*
Copyright 2023.

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

	v1alpha1 "github.com/tsuru/acl-operator/api/v1alpha1"
	aclapi "github.com/tsuru/acl-operator/clients/aclapi"
	batchv1 "k8s.io/api/batch/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	tsuruJobLabel     = "tsuru.io/job-name"
	tsuruJobACLPrefix = "tsuru-job-"
)

// TsuruAppReconciler reconciles a Tsuru App object
type TsuruCronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ACLAPI aclapi.Client
}

func (r *TsuruCronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	job := &batchv1.CronJob{}
	err := r.Client.Get(ctx, req.NamespacedName, job)
	if err != nil {
		l.Error(err, "could not get CronJob object")
		return ctrl.Result{}, err
	}

	jobName := job.Labels[tsuruJobLabel]

	if jobName == "" {
		return ctrl.Result{}, nil
	}

	rules, err := r.ACLAPI.JobRules(ctx, jobName)
	if err != nil {
		l.Error(err, "could not get Tsuru Job Rules from ACLAPI")
		return ctrl.Result{}, err
	}

	destinations, errs := convertACLAPIRulesToOperatorRules(rules)

	warningErrors := []string{}
	for _, e := range errs {
		warningErrors = append(warningErrors, e.Error())
	}

	aclName := tsuruJobACLPrefix + jobName

	acl := &v1alpha1.ACL{}
	err = r.Client.Get(ctx, client.ObjectKey{
		Name:      aclName,
		Namespace: job.Namespace,
	}, acl)

	if k8sErrors.IsNotFound(err) {
		if len(destinations) == 0 {
			return ctrl.Result{}, nil
		}

		err = r.Client.Create(ctx, &v1alpha1.ACL{
			ObjectMeta: metav1.ObjectMeta{
				Name:      aclName,
				Namespace: job.Namespace,
			},
			Spec: v1alpha1.ACLSpec{
				Source: v1alpha1.ACLSpecSource{
					TsuruJob: jobName,
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

	acl.Spec.Source = v1alpha1.ACLSpecSource{
		TsuruJob: jobName,
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

// SetupWithManager sets up the controller with the Manager.
func (r *TsuruCronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.CronJob{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2, RecoverPanic: true}).
		Complete(r)
}
