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

package main

import (
	"flag"
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/tsuru/acl-operator/api/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/tsuru/acl-operator/clients/aclapi"
	"github.com/tsuru/acl-operator/clients/tsuruapi"
	"github.com/tsuru/acl-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	var aclAPIAddr string
	var aclAPIUser string
	var aclAPIPassword string

	var tsuruAPIAddr string
	var tsuruAPIToken string

	flag.StringVar(&aclAPIAddr, "acl-api-address", "", "The address of ACL API [required]")
	flag.StringVar(&aclAPIUser, "acl-api-user", "", "The user of ACL API [required]")
	flag.StringVar(&aclAPIPassword, "acl-api-password", "", "The password of ACL API [required]")

	flag.StringVar(&tsuruAPIAddr, "tsuru-api-address", "", "The address of Tsuru API [required]")
	flag.StringVar(&tsuruAPIToken, "tsuru-api-token", "", "The token of Tsuru API [required")

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development:     true,
		StacktraceLevel: zapcore.DPanicLevel,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))

	ctrl.SetLogger(logger)

	tsuruAppReconciler := true
	if aclAPIAddr == "" || aclAPIUser == "" || aclAPIPassword == "" {
		logger.Info("TsuruAppReconciler is disabled due a missing acl api settings")
		tsuruAppReconciler = false
	}

	if tsuruAPIAddr == "" {
		tsuruAPIAddr = os.Getenv("TSURU_TARGET")
	}

	if tsuruAPIToken == "" {
		tsuruAPIToken = os.Getenv("TSURU_TOKEN")
	}

	if tsuruAPIAddr == "" {
		fmt.Println("TSURU_TARGET env or tsuru-api-address flag is not defined")
		os.Exit(1)
	}

	if tsuruAPIToken == "" {
		fmt.Println("TSURU_TOKEN env or tsuru-api-token flag is not defined")
		os.Exit(1)
	}

	tsuruAPI := tsuruapi.New(tsuruAPIAddr, tsuruAPIToken)
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme.Scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e1803b8b.extensions.tsuru.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ACLReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		TsuruAPI: tsuruAPI,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ACL")
		os.Exit(1)
	}
	if err = (&controllers.ACLDNSEntryReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Resolver: controllers.DefaultResolver,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ACLDNSEntry")
		os.Exit(1)
	}

	if tsuruAppReconciler {
		if err = (&controllers.TsuruAppReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			ACLAPI: aclapi.New(aclAPIAddr, aclAPIUser, aclAPIPassword),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "TsuruAppReconciler")
			os.Exit(1)
		}
	}

	if err = (&controllers.RpaasInstanceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RpaasInstanceReconciler")
		os.Exit(1)
	}

	if err = (&controllers.TsuruAppAdressReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Resolver: controllers.DefaultResolver,
		TsuruAPI: tsuruAPI,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TsuruAppAdress")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
