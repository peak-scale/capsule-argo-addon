package main

import (
	"log"
	"os"

	"git.bedag.cloud/gelan/gelan-infra/controllers/tenancy-controller/pkg/controller"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	crlog "sigs.k8s.io/controller-runtime/pkg/log"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type rootCmdFlags struct {
	logger                       logr.Logger
	logLevel                     int
	capsuleProxyServiceName      string
	capsuleProxyServiceNamespace string
	capsuleProxyServicePort      int32
	userTenantNamespace          string
	systemTenantNamespace        string
	enableLeaderElection         bool
	metricsAddr                  string
	argoCDNamespace              string
}

var (
	scheme     = runtime.NewScheme()
	setupLog   = ctrl.Log.WithName("setup")
	rootLogger = stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{LogCaller: stdr.All})
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(capsulev1beta2.AddToScheme(scheme))
}

func main() {
	options := rootCmdFlags{
		logger:                       rootLogger.WithName("main"),
		capsuleProxyServiceName:      "capsule-proxy",
		capsuleProxyServiceNamespace: "capsule-system",
		userTenantNamespace:          "tenants",
		systemTenantNamespace:        "tenants-system",
		capsuleProxyServicePort:      9001,
		argoCDNamespace:              "argocd",
		logLevel:                     3,
	}

	crlog.SetLogger(rootLogger.WithName("controller-runtime"))

	rootCommand := cobra.Command{
		Use: "tunnel-controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			stdr.SetVerbosity(options.logLevel)
			logger := options.logger
			logger.Info("logging verbosity", "level", options.logLevel)

			manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
				Scheme: scheme,
				Metrics: metricsserver.Options{
					BindAddress: options.metricsAddr,
				},
				LeaderElection:         options.enableLeaderElection,
				LeaderElectionID:       "2cadwd3jea.gelan.cloud",
				HealthProbeBindAddress: ":10080",
				NewClient: func(config *rest.Config, options client.Options) (client.Client, error) {
					options.Cache.Unstructured = true

					return client.New(config, options)
				},
			})
			if err != nil {
				logger.Error(err, "unable to start manager")
				os.Exit(1)
			}

			_ = manager.AddReadyzCheck("ping", healthz.Ping)
			_ = manager.AddHealthzCheck("ping", healthz.Ping)

			ctx := ctrl.SetupSignalHandler()

			if err = (&controller.TenancyController{
				Client:   manager.GetClient(),
				Log:      ctrl.Log.WithName("controllers").WithName("Tenant"),
				Recorder: manager.GetEventRecorderFor("tenancy-controller"),
				Options: controller.TenancyControllerOptions{
					CapsuleProxyServiceName:      options.capsuleProxyServiceName,
					CapsuleProxyServiceNamespace: options.capsuleProxyServiceNamespace,
					CapsuleProxyServicePort:      options.capsuleProxyServicePort,
					UserTenantNamespace:          options.userTenantNamespace,
					SystemTenantNamespace:        options.systemTenantNamespace,
					ArgoCDNamespace:              options.argoCDNamespace,
				},
			}).SetupWithManager(ctx, manager); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "Tenant")
				os.Exit(1)
			}

			setupLog.Info("propagation manager start serving")

			if err = manager.Start(ctx); err != nil {
				setupLog.Error(err, "problem running manager")
				os.Exit(1)
			}

			return nil
		},
	}

	rootCommand.PersistentFlags().StringVar(&options.capsuleProxyServiceName, "proxy-svc-name", options.capsuleProxyServiceName, "capsule proxy service name")
	rootCommand.PersistentFlags().StringVar(&options.capsuleProxyServiceNamespace, "proxy-svc-namespace", options.capsuleProxyServiceNamespace, "capsule proxy serice namespace")
	rootCommand.PersistentFlags().IntVarP(&options.logLevel, "log-level", "v", options.logLevel, "numeric log level")
	rootCommand.PersistentFlags().StringVar(&options.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	rootCommand.PersistentFlags().BoolVar(&options.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	err := rootCommand.Execute()
	if err != nil {
		panic(err)
	}
}
