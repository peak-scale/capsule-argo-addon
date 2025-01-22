// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "github.com/KimMachineGun/automemlimit"
	_ "go.uber.org/automaxprocs"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	capsuleindexer "github.com/projectcapsule/capsule/pkg/indexer"
	tntindex "github.com/projectcapsule/capsule/pkg/indexer/tenant"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/config"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/tenant"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
	"github.com/peak-scale/capsule-argo-addon/internal/metrics"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	"github.com/peak-scale/capsule-argo-addon/internal/webhooks"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(capsulev1beta2.AddToScheme(scheme))
	utilruntime.Must(argocdv1alpha1.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection, enablePprof, hooks bool
	var probeAddr string
	var settingName string

	ctx := ctrl.SetupSignalHandler()

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&settingName, "setting-name", "default", "The setting name to use for this controller instance")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":10080", "The address the probe endpoint binds to.")
	flag.BoolVar(&enablePprof, "enable-pprof", false, "Enables Pprof endpoint for profiling (not recommend in production)")
	flag.BoolVar(&hooks, "enable-webhooks", false, "Register Mutating Webhooks to be serving")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrlConfig := ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "fefb3d10.projectcapsule.dev",
		PprofBindAddress:       ":8082",
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
		LeaderElectionReleaseOnCancel: true,
	}

	// Conditional config
	if hooks {
		ctrlConfig.WebhookServer = ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port: 9443,
		})
	}

	if enablePprof {
		ctrlConfig.PprofBindAddress = ":8082"
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrlConfig)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if hooks {
		setupLog.Info("registering webhooks")

		mgr.GetWebhookServer().Register("/mutate/applications", &admission.Webhook{
			Handler: &webhooks.ApplicationWebhook{
				Decoder: admission.NewDecoder(mgr.GetScheme()),
				Client:  mgr.GetClient(),
				Log:     ctrl.Log.WithName("Webhooks").WithName("Applications"),
			},
		})
		mgr.GetWebhookServer().Register("/mutate/applicationsets", &admission.Webhook{
			Handler: &webhooks.ApplicationSetWebhook{
				Decoder: admission.NewDecoder(mgr.GetScheme()),
				Client:  mgr.GetClient(),
				Log:     ctrl.Log.WithName("Webhooks").WithName("ApplicationSets"),
			},
		})
	}

	// Indexer
	indexers := []capsuleindexer.CustomIndexer{
		&tntindex.NamespacesReference{Obj: &capsulev1beta2.Tenant{}},
	}

	for _, fieldIndex := range indexers {
		if err = mgr.GetFieldIndexer().IndexField(
			ctx,
			fieldIndex.Object(),
			fieldIndex.Field(),
			fieldIndex.Func(),
		); err != nil {
			setupLog.Error(err, "cannot create new Field Indexer")
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder
	store := stores.NewConfigStore()

	metricsRecorder := metrics.MustMakeRecorder()

	settings := &config.ConfigReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Log:      ctrl.Log.WithName("Controllers").WithName("Config"),
		Recorder: mgr.GetEventRecorderFor("config-controller"),
		Store:    store,
		Config: config.ReconcilerConfig{
			SettingName: settingName,
		},
	}

	directClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: mgr.GetScheme(),
		Mapper: mgr.GetRESTMapper(),
	})
	if err != nil {
		setupLog.Error(err, "unable to initialize client")
		os.Exit(1)
	}

	if err = settings.Initialize(ctx, directClient); err != nil {
		setupLog.Error(err, "unable to initialize settings")
		os.Exit(1)
	}
	if err := settings.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Config")
		os.Exit(1)
	}

	if err = (&tenant.TenancyController{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("Controllers").WithName("Tenant"),
		Recorder: mgr.GetEventRecorderFor("tenant-controller"),
		Scheme:   mgr.GetScheme(),
		Metrics:  metricsRecorder,
		Settings: store,
		Rest:     mgr.GetConfig(),
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenant")
		os.Exit(1)
	}
	setupLog.Info("tenant-controller initialized")

	if err = (&translator.TranslatorController{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("Controllers").WithName("Translator"),
		Recorder: mgr.GetEventRecorderFor("translator-controller"),
		Scheme:   mgr.GetScheme(),
		Settings: store,
	}).SetupWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Translator")
		os.Exit(1)
	}
	setupLog.Info("translator-controller initialized")

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
