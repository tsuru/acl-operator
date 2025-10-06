package controllers

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-logr/logr"
	"github.com/tsuru/acl-operator/api/v1alpha1"
	tsuruv1 "github.com/tsuru/tsuru/provision/kubernetes/pkg/apis/tsuru/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ACLGarbageCollector struct {
	client.Client
	DryRun       bool
	DryRunOutput io.Writer
	Logger       logr.Logger
}

type appACLKey struct {
	App       string
	Namespace string
}

type jobACLKey struct {
	Job       string
	Namespace string
}

func (a *ACLGarbageCollector) Run(ctx context.Context) {
	time.Sleep(time.Second * 30) // wait for sync
	for {
		err := a.Loop(ctx)
		if err != nil {
			fmt.Println("Err loop: ", err)
		}

		time.Sleep(time.Minute * 5)
	}
}

func (a *ACLGarbageCollector) Loop(ctx context.Context) error {
	jobACLs := map[jobACLKey]struct{}{}

	allDNSEntries, err := a.allDNSEntries(ctx)
	if err != nil {
		return err
	}
	dnsEntries := make(map[string]struct{}, len(allDNSEntries))
	for _, dnsEntry := range allDNSEntries {
		dnsEntries[dnsEntry.Spec.Host] = struct{}{}
	}

	allTsuruAppAddress, err := a.allTsuruAppAddress(ctx)
	if err != nil {
		return err
	}
	tsuruApps := make(map[string]struct{}, len(allTsuruAppAddress))
	appACLs := make(map[appACLKey]struct{}, len(allTsuruAppAddress)) // fair aproximation
	for _, tsuruAppAddress := range allTsuruAppAddress {
		tsuruApps[tsuruAppAddress.Spec.Name] = struct{}{}
	}

	allRPaaSInstancesAddresses, err := a.allRPaaSInstancesAddresses(ctx)
	if err != nil {
		return err
	}
	rpaaInstances := make(map[v1alpha1.ACLSpecRpaasInstance]string, len(allRPaaSInstancesAddresses))
	for _, rpaaInstanceAddress := range allRPaaSInstancesAddresses {
		key := v1alpha1.ACLSpecRpaasInstance{
			ServiceName: rpaaInstanceAddress.Spec.ServiceName,
			Instance:    rpaaInstanceAddress.Spec.Instance,
		}

		rpaaInstances[key] = rpaaInstanceAddress.Name
	}

	allACLSs, err := a.allACLs(ctx)
	if err != nil {
		return err
	}
	for _, acl := range allACLSs {
		if acl.Spec.Source.TsuruApp != "" {
			appACLs[appACLKey{
				Namespace: acl.Namespace,
				App:       acl.Spec.Source.TsuruApp,
			}] = struct{}{}
		}

		if acl.Spec.Source.TsuruJob != "" {
			jobACLs[jobACLKey{
				Namespace: acl.Namespace,
				Job:       acl.Spec.Source.TsuruJob,
			}] = struct{}{}
		}

		for _, destination := range acl.Spec.Destinations {
			if destination.ExternalDNS != nil {
				_, found := dnsEntries[destination.ExternalDNS.Name]
				if found {
					delete(dnsEntries, destination.ExternalDNS.Name) // the remain keys on dnsEntries must be garbage collected
				}
			} else if destination.TsuruApp != "" {
				_, found := tsuruApps[destination.TsuruApp]
				if found {
					delete(tsuruApps, destination.TsuruApp) // the remain keys on tsuruApps must be garbage collected
				}
			} else if destination.RpaasInstance != nil {
				_, found := rpaaInstances[*destination.RpaasInstance]
				if found {
					delete(rpaaInstances, *destination.RpaasInstance) // the remain keys on rpaaInstances must be garbage collected
				}
			}
		}
	}

	allTsuruApps, err := a.allTsuruApps(ctx)
	if err != nil {
		return err
	}
	for _, tsuruApp := range allTsuruApps {
		key := appACLKey{
			App:       tsuruApp.Name,
			Namespace: tsuruApp.Spec.NamespaceName,
		}
		_, found := appACLs[key]
		if found {
			delete(appACLs, key) // the remain keys on appACLs must be garbage collected
		}
	}

	allTsuruJobs, err := a.allTsuruJobs(ctx)
	if err != nil {
		return err
	}
	for _, tsuruJob := range allTsuruJobs {
		key := jobACLKey{
			Job:       tsuruJob.Labels[tsuruJobLabel],
			Namespace: tsuruJob.Namespace,
		}
		_, found := jobACLs[key]
		if found {
			delete(jobACLs, key) // the remain keys on appACLs must be garbage collected
		}
	}

	if a.DryRun {
		for dnsEntry := range dnsEntries {
			fmt.Fprintln(a.DryRunOutput, "dnsEntry is marked to delete", dnsEntry)
		}
		for tsuruApp := range tsuruApps {
			fmt.Fprintf(a.DryRunOutput, "tsuruApp is marked to delete: %q\n", tsuruApp)
		}

		for _, rpaasInstanceName := range rpaaInstances {
			fmt.Fprintln(a.DryRunOutput, "rpaaInstance is marked to delete", rpaasInstanceName)
		}

		for appACL := range appACLs {
			fmt.Fprintln(a.DryRunOutput, "APP ACL is marked to delete", appACL.Namespace, "/", appACL.App)
		}

		for jobACL := range jobACLs {
			fmt.Fprintln(a.DryRunOutput, "Job ACL is marked to delete", jobACL.Namespace, "/", jobACL.Job)
		}
		return nil
	}

	for dnsEntry := range dnsEntries {
		err = a.Delete(ctx, &v1alpha1.ACLDNSEntry{
			ObjectMeta: v1.ObjectMeta{
				Name: validResourceName(dnsEntry),
			},
		})
		if err != nil {
			a.Logger.Error(err, "failed to remove dnsEntry", "dnsEntry", dnsEntry)
		}
	}

	for tsuruApp := range tsuruApps {
		err = a.Delete(ctx, &v1alpha1.TsuruAppAddress{
			ObjectMeta: v1.ObjectMeta{
				Name: validResourceName(tsuruApp),
			},
		})
		if err != nil {
			a.Logger.Error(err, "failed to remove tsuruAppAddress", "tsuruApp", tsuruApp)
		}
	}

	for _, rpaasInstanceName := range rpaaInstances {
		err = a.Delete(ctx, &v1alpha1.RpaasInstanceAddress{
			ObjectMeta: v1.ObjectMeta{
				Name: rpaasInstanceName,
			},
		})
		if err != nil {
			a.Logger.Error(err, "failed to remove rpaasInstanceAddress", "rpaasInstanceAddress", rpaasInstanceName)
		}
	}

	for appACL := range appACLs {
		err = a.Delete(ctx, &v1alpha1.ACL{
			ObjectMeta: v1.ObjectMeta{
				Namespace: appACL.Namespace,
				Name:      appACL.App,
			},
		})
		if err != nil {
			a.Logger.Error(err, "failed to remove acl", "namespace", appACL.Namespace, "name", appACL.App)
		}
	}

	for jobACL := range jobACLs {
		err = a.Delete(ctx, &v1alpha1.ACL{
			ObjectMeta: v1.ObjectMeta{
				Namespace: jobACL.Namespace,
				Name:      tsuruJobACLPrefix + jobACL.Job,
			},
		})
		if err != nil {
			a.Logger.Error(err, "failed to remove acl", "namespace", jobACL.Namespace, "name", jobACL.Job)
		}
	}

	return nil
}

func (a *ACLGarbageCollector) allACLs(ctx context.Context) ([]v1alpha1.ACL, error) {
	result := []v1alpha1.ACL{}

	continueToken := ""

	for {
		allACLSs := &v1alpha1.ACLList{}

		err := a.List(ctx, allACLSs, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, allACLSs.Items...)

		if allACLSs.Continue == "" {
			break
		}

		continueToken = allACLSs.Continue
	}

	return result, nil
}

func (a *ACLGarbageCollector) allDNSEntries(ctx context.Context) ([]v1alpha1.ACLDNSEntry, error) {
	result := []v1alpha1.ACLDNSEntry{}

	continueToken := ""

	for {
		allDNSEntries := &v1alpha1.ACLDNSEntryList{}

		err := a.List(ctx, allDNSEntries, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, allDNSEntries.Items...)

		if allDNSEntries.Continue == "" {
			break
		}

		continueToken = allDNSEntries.Continue
	}

	return result, nil
}

func (a *ACLGarbageCollector) allTsuruAppAddress(ctx context.Context) ([]v1alpha1.TsuruAppAddress, error) {
	result := []v1alpha1.TsuruAppAddress{}

	continueToken := ""

	for {
		allTsuruAppAddress := &v1alpha1.TsuruAppAddressList{}

		err := a.List(ctx, allTsuruAppAddress, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, allTsuruAppAddress.Items...)

		if allTsuruAppAddress.Continue == "" {
			break
		}

		continueToken = allTsuruAppAddress.Continue
	}

	return result, nil
}

func (a *ACLGarbageCollector) allRPaaSInstancesAddresses(ctx context.Context) ([]v1alpha1.RpaasInstanceAddress, error) {
	result := []v1alpha1.RpaasInstanceAddress{}

	continueToken := ""

	for {
		allRPaaSInstancesAddress := &v1alpha1.RpaasInstanceAddressList{}

		err := a.List(ctx, allRPaaSInstancesAddress, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, allRPaaSInstancesAddress.Items...)

		if allRPaaSInstancesAddress.Continue == "" {
			break
		}

		continueToken = allRPaaSInstancesAddress.Continue
	}

	return result, nil
}

func (a *ACLGarbageCollector) allTsuruApps(ctx context.Context) ([]tsuruv1.App, error) {
	result := []tsuruv1.App{}

	continueToken := ""

	for {
		allTsuruApps := &tsuruv1.AppList{}

		err := a.List(ctx, allTsuruApps, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}
		result = append(result, allTsuruApps.Items...)

		if allTsuruApps.Continue == "" {
			break
		}

		continueToken = allTsuruApps.Continue
	}

	return result, nil
}

func (a *ACLGarbageCollector) allTsuruJobs(ctx context.Context) ([]batchv1.CronJob, error) {
	result := []batchv1.CronJob{}

	continueToken := ""

	for {
		allTsuruJobs := &batchv1.CronJobList{}

		err := a.List(ctx, allTsuruJobs, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return nil, err
		}

	cronjobLoop:
		for _, cronjob := range allTsuruJobs.Items {
			if cronjob.Labels[tsuruJobLabel] == "" {
				continue cronjobLoop
			}
			result = append(result, cronjob)
		}

		if allTsuruJobs.Continue == "" {
			break
		}

		continueToken = allTsuruJobs.Continue
	}

	return result, nil
}
