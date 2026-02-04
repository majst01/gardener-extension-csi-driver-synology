package healthcheck

import (
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck"
	"github.com/gardener/gardener/extensions/pkg/controller/healthcheck/general"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// DefaultAddOptions are the default AddOptions for AddToManager
var DefaultAddOptions = AddOptions{}

// AddOptions are options to apply when adding the healthcheck controller to the manager
type AddOptions struct {
	// Controller are the controller related options
	Controller healthcheck.ControllerOptions
}

// AddToManager adds a controller with the default Options
func AddToManager(mgr manager.Manager) error {
	return AddToManagerWithOptions(mgr, DefaultAddOptions)
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager
func AddToManagerWithOptions(mgr manager.Manager, opts AddOptions) error {
	opts.Controller.Name = "healthcheck-" + constants.ExtensionType
	opts.Controller.Type = constants.ExtensionType

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
				HealthCheck:     general.CheckManagedResource(constants.NodeName),
			},
			{
				ConditionType: "DeploymentHealthy",
				HealthCheck: general.NewDeploymentChecker(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      constants.ControllerName,
						Namespace: v1beta1constants.GardenNamespace,
					},
				}),
			},
			{
				ConditionType: "DaemonSetHealthy",
				HealthCheck: general.NewDaemonSetChecker(&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      constants.NodeName,
						Namespace: v1beta1constants.GardenNamespace,
					},
				}),
			},
		},
	).AddToManager(mgr, opts.Controller)
}
