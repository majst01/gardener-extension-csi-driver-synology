package app

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	componentbaseconfig "k8s.io/component-base/config/v1alpha1"

	heartbeatcmd "github.com/gardener/gardener/extensions/pkg/controller/heartbeat/cmd"
	"github.com/gardener/gardener/extensions/pkg/util"
	"github.com/gardener/gardener/pkg/apis/authentication/install"
	"github.com/labstack/gommon/log"
	csidriversynologycmd "github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/cmd"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	heartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"
	ghealth "github.com/gardener/gardener/pkg/healthz"

	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/controller/lifecycle"
)

type Options struct {
	generalOptions     *controllercmd.GeneralOptions
	configOptions      *csidriversynologycmd.ConfigOptions
	restOptions        *controllercmd.RESTOptions
	managerOptions     *controllercmd.ManagerOptions
	controllerOptions  *controllercmd.ControllerOptions
	heartbeatOptions   *heartbeatcmd.Options
	healthOptions      *controllercmd.ControllerOptions
	controllerSwitches *controllercmd.SwitchOptions
	reconcileOptions   *controllercmd.ReconcilerOptions
	optionAggregator   controllercmd.OptionAggregator
}

func NewOptions() *Options {
	options := &Options{
		generalOptions: &controllercmd.GeneralOptions{},
		configOptions:  &csidriversynologycmd.ConfigOptions{},
		restOptions:    &controllercmd.RESTOptions{},
		managerOptions: &controllercmd.ManagerOptions{
			LeaderElection:          true,
			LeaderElectionID:        controllercmd.LeaderElectionNameID(constants.ExtensionName),
			LeaderElectionNamespace: os.Getenv("LEADER_ELECTION_NAMESPACE"),
			MetricsBindAddress:      ":8080",
			HealthBindAddress:       ":8081",
		},

		// options for the controlplane controller
		controllerOptions: &controllercmd.ControllerOptions{
			MaxConcurrentReconciles: 5,
		},

		heartbeatOptions: &heartbeatcmd.Options{
			// This is a default value.
			ExtensionName:        constants.ExtensionName,
			RenewIntervalSeconds: 30,
			Namespace:            os.Getenv("LEADER_ELECTION_NAMESPACE"),
		},
		healthOptions: &controllercmd.ControllerOptions{
			// This is a default value.
			MaxConcurrentReconciles: 5,
		},
		controllerSwitches: csidriversynologycmd.ControllerSwitchOptions(),
		reconcileOptions:   &controllercmd.ReconcilerOptions{},
	}

	options.optionAggregator = controllercmd.NewOptionAggregator(
		options.generalOptions,
		// options.csidriverlvmOptions,
		options.restOptions,
		options.managerOptions,
		options.controllerOptions,
		controllercmd.PrefixOption("heartbeat-", options.heartbeatOptions),
		controllercmd.PrefixOption("healthcheck-", options.healthOptions),
		options.controllerSwitches,
		options.reconcileOptions,
	)
	return options
}

func (options *Options) run(ctx context.Context) error {
	log.Info("starting " + constants.ExtensionName)

	util.ApplyClientConnectionConfigurationToRESTConfig(&componentbaseconfig.ClientConnectionConfiguration{
		QPS:   100.0,
		Burst: 130,
	}, options.restOptions.Completed().Config)

	log.Info("applied rest config")

	mgrOpts := options.managerOptions.Completed().Options()

	log.Info("completed mgr-options")

	mgrOpts.Client = client.Options{
		Cache: &client.CacheOptions{
			DisableFor: []client.Object{
				&corev1.Secret{},
				&corev1.ConfigMap{},
			},
		},
	}

	mgr, err := manager.New(options.restOptions.Completed().Config, mgrOpts)
	if err != nil {
		return fmt.Errorf("could not instantiate controller-manager: %w", err)
	}
	log.Info("completed rest-options")

	err = extensionscontroller.AddToScheme(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("could not add mgr-scheme to extension-controller: %w", err)
	}
	log.Info("added mgr-scheme to extensionscontroller")

	err = install.AddToScheme(mgr.GetScheme())
	if err != nil {
		return fmt.Errorf("could not add mgr-scheme to installation")
	}
	log.Info("added mgr-scheme to installation")

	ctrlConfig := options.configOptions.Completed()
	ctrlConfig.Apply(&lifecycle.DefaultAddOptions.Config)

	options.controllerOptions.Completed().Apply(&lifecycle.DefaultAddOptions.ControllerOptions)
	options.reconcileOptions.Completed().Apply(&lifecycle.DefaultAddOptions.IgnoreOperationAnnotation, &lifecycle.DefaultAddOptions.ExtensionClass)
	options.heartbeatOptions.Completed().Apply(&heartbeatcontroller.DefaultAddOptions)

	if err := options.controllerSwitches.Completed().AddToManager(ctx, mgr); err != nil {
		return fmt.Errorf("could not add controllers to manager: %w", err)
	}
	log.Info("added controllers to manager")

	if err := mgr.AddReadyzCheck("informer-sync", ghealth.NewCacheSyncHealthz(mgr.GetCache())); err != nil {
		return fmt.Errorf("could not add ready check for informers: %w", err)
	}
	log.Info("added readyzcheck")

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		return fmt.Errorf("could not add health check to manager: %w", err)
	}
	log.Info("added healthzcheck")

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}

	return nil
}
