package lifecycle

import (
	"context"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/apis/config"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// DefaultAddOptions are the default AddOptions for AddToManager
var DefaultAddOptions = AddOptions{}

// AddOptions are options to apply when adding the extension controller to the manager
type AddOptions struct {
	// Controller are the controller related options
	ControllerOptions controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not
	IgnoreOperationAnnotation bool
	// ExtensionClass defines the extension class this extension is responsible for
	ExtensionClass extensionsv1alpha1.ExtensionClass
	// Config is the extension configuration
	Config config.ControllerConfiguration
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager
func AddToManagerWithOptions(mgr manager.Manager, opts AddOptions) error {
	return extension.Add(mgr, extension.AddArgs{
		Actuator:          NewActuator(mgr.GetClient(), &opts.Config),
		ControllerOptions: opts.ControllerOptions,
		Name:              constants.ExtensionName,
		FinalizerSuffix:   constants.ExtensionType,
		Resync:            0,
		Predicates:        extension.DefaultPredicates(context.Background(), mgr, opts.IgnoreOperationAnnotation),
		Type:              constants.ExtensionType,
	})
}

// AddToManager adds a controller with the default Options
func AddToManager(ctx context.Context, mgr manager.Manager) error {
	return AddToManagerWithOptions(mgr, DefaultAddOptions)
}
