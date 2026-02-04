package healthcheck

import (
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/general"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// DefaultAddArgs are the default DefaultAddArgs for AddToManager
var DefaultAddArgs = healthcheck.DefaultAddArgs{}

// AddToManager adds a controller with the default options
func AddToManager(mgr manager.Manager) error {
	return AddToManagerWithOptions(mgr, DefaultAddArgs)
}

// AddToManagerWithOptions adds a controller with the given options to the given manager
func AddToManagerWithOptions(mgr manager.Manager, opts healthcheck.DefaultAddArgs) error {
	return healthcheck.DefaultRegistration(
		constants.ExtensionType,
		extensionsv1alpha1.SchemeGroupVersion.WithKind(extensionsv1alpha1.WorkerResource),
		func() client.ObjectList { return &extensionsv1alpha1.WorkerList{} },
		func() extensionsv1alpha1.Object { return &extensionsv1alpha1.Worker{} },
		mgr,
		opts,
		nil,
		[]healthcheck.ConditionTypeToHealthCheck{
			{
				ConditionType: "ControllerHealthy",
				HealthCheck:   general.CheckManagedResource(constants.ControllerName),
			},
			{
				ConditionType: "NodeHealthy",
				HealthCheck:   general.CheckManagedResource(constants.NodeName),
			},
			{
				ConditionType: "DeploymentHealthy",
				HealthCheck:   general.NewSeedDeploymentHealthChecker(constants.ControllerName),
			},
			{
				ConditionType: "DaemonSetHealthy",
				HealthCheck:   general.NewSeedDaemonSetHealthChecker(constants.NodeName),
			},
		},
		sets.New[gardencorev1beta1.ConditionType](),
	)
}
