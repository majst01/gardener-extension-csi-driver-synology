package lifecycle

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/utils/managedresources"
	"github.com/go-logr/logr"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/apis/config"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/apis/csidriversynology/v1alpha1"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/synology"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Actuator acts upon Extension resources
type Actuator struct {
	client  client.Client
	decoder runtime.Decoder
	config  config.ControllerConfiguration
}

// NewActuator creates a new Actuator
func NewActuator(client client.Client, config config.ControllerConfiguration) extension.Actuator {
	return &Actuator{
		client:  client,
		decoder: serializer.NewCodecFactory(client.Scheme(), serializer.EnableStrict).UniversalDecoder(),
		config:  config,
	}
}

// Reconcile the Extension resource
func (a *Actuator) Reconcile(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	shootConfig := &v1alpha1.CsiDriverSynologyConfig{}
	if ex.Spec.ProviderConfig != nil {
		_, _, err := a.decoder.Decode(ex.Spec.ProviderConfig.Raw, nil, shootConfig)
		if err != nil {
			return fmt.Errorf("failed to decode provider config: %w", err)
		}
	}

	namespace := ex.GetNamespace()
	shootName := namespace
	shootNamespace := namespace

	log.Info("Reconciling Synology CSI extension", "namespace", namespace)

	// Create Synology client
	synologyClient, err := synology.NewClient(
		a.config.SynologyURL,
		a.config.AdminUsername,
		a.config.AdminPassword,
	)

	if err != nil {
		return fmt.Errorf("failed to create Synology client: %w", err)
	}

	// Login to Synology
	if err := synologyClient.Login(); err != nil {
		return fmt.Errorf("failed to login to Synology NAS: %w", err)
	}
	defer synologyClient.Logout()

	// Generate credentials for this shoot
	shootUsername := synology.GenerateShootUsername(shootName, shootNamespace)
	shootPassword, err := synology.GenerateRandomPassword(16)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	// Create user on Synology
	if err := synologyClient.CreateUser(shootUsername, shootPassword); err != nil {
		return fmt.Errorf("failed to create user on Synology: %w", err)
	}

	// Generate CHAP credentials if enabled
	var chapUsername, chapPassword string
	if a.config.ChapEnabled {
		chapUsername = shootUsername + "-chap"
		chapPassword, err = synology.GenerateRandomPassword(16)
		if err != nil {
			return fmt.Errorf("failed to generate CHAP password: %w", err)
		}
	}

	u, err := url.Parse(a.config.SynologyURL)
	if err != nil {
		return fmt.Errorf("failed to parse synology-url: %w", err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return fmt.Errorf("failed to parse synology-url port: %w", err)
	}

	// Create manifest config
	manifestConfig := &synology.ManifestConfig{
		Namespace:    "kube-system",
		Url:          a.config.SynologyURL,
		Username:     shootUsername,
		Password:     shootPassword,
		ChapEnabled:  a.config.ChapEnabled,
		ChapUsername: chapUsername,
		ChapPassword: chapPassword,
		Clients: []synology.ClientConfig{
			{
				Host:     u.Hostname(),
				Port:     port,
				HTTPS:    u.Scheme == "https",
				Username: shootUsername,
				Password: shootPassword,
			},
			{
				Host:     u.Hostname(),
				Port:     5001,
				HTTPS:    u.Scheme == "https",
				Username: shootUsername,
				Password: shootPassword,
			},
		},
	}

	objects, err := a.generateManifests(manifestConfig)
	if err != nil {
		return fmt.Errorf("unable to generate resource manifests for shoot: %w", err)
	}

	shootResources, err := managedresources.NewRegistry(kubernetes.ShootScheme, kubernetes.ShootCodec, kubernetes.ShootSerializer).AddAllAndSerialize(objects...)
	if err != nil {
		return fmt.Errorf("unable to create registry: %w", err)
	}

	err = managedresources.CreateForShoot(ctx, a.client, ex.Namespace, constants.CSIDriverName, constants.ExtensionType, false, shootResources)
	if err != nil {
		return fmt.Errorf("unable to create shoot resources: %w", err)
	}

	log.Info("Successfully reconciled Synology CSI extension")
	return nil
}

// Delete the Extension resource
func (a *Actuator) Delete(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	namespace := ex.GetNamespace()
	shootName := namespace
	shootNamespace := namespace

	log.Info("Deleting Synology CSI extension", "namespace", namespace)

	// Create Synology client
	synologyClient, err := synology.NewClient(
		a.config.SynologyURL,
		a.config.AdminUsername,
		a.config.AdminPassword,
	)

	if err != nil {
		return fmt.Errorf("unable to create Synology client: %w", err)
	}

	// Login to Synology
	if err := synologyClient.Login(); err != nil {
		log.Error(err, "Failed to login to Synology NAS, continuing with resource deletion")
	} else {
		defer synologyClient.Logout()

		// Delete user from Synology
		shootUsername := synology.GenerateShootUsername(shootName, shootNamespace)
		if err := synologyClient.DeleteUser(shootUsername); err != nil {
			log.Error(err, "Failed to delete user from Synology", "username", shootUsername)
		}
	}

	// Delete resources from shoot cluster
	if err := a.deleteResources(ctx, log, v1beta1constants.GardenNamespace); err != nil {
		return fmt.Errorf("failed to delete resources: %w", err)
	}

	log.Info("Successfully deleted Synology CSI extension")
	return nil
}

// Restore the Extension resource
func (a *Actuator) Restore(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Reconcile(ctx, log, ex)
}

// Migrate the Extension resource
func (a *Actuator) Migrate(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return nil
}

// ForceDelete forcefully deletes the Extension resource
func (a *Actuator) ForceDelete(ctx context.Context, log logr.Logger, ex *extensionsv1alpha1.Extension) error {
	return a.Delete(ctx, log, ex)
}

// generateManifests deploys all necessary resources to the shoot cluster
func (a *Actuator) generateManifests(config *synology.ManifestConfig) ([]client.Object, error) {
	secret, err := synology.GenerateSecret(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	objects := []client.Object{
		synology.GenerateServiceAccount(config.Namespace, constants.ControllerName),
		synology.GenerateServiceAccount(config.Namespace, constants.NodeName),
		synology.GenerateControllerClusterRole(),
		synology.GenerateNodeClusterRole(),
		synology.GenerateClusterRoleBinding(constants.ControllerName, config.Namespace, constants.ControllerName),
		synology.GenerateClusterRoleBinding(constants.NodeName, config.Namespace, constants.NodeName),
		secret,
		synology.GenerateCSIDriver(),
		synology.GenerateService(config.Namespace),
		synology.GenerateControllerDeployment(config.Namespace),
		synology.GenerateNodeDaemonSet(config.Namespace),
		synology.GenerateStorageClass(config.Namespace),
		synology.GenerateAllowAllEgressNetworkPolicy(config.Namespace),
	}

	return objects, nil
}

// deleteResources deletes all resources from the shoot cluster
func (a *Actuator) deleteResources(ctx context.Context, log logr.Logger, namespace string) error {
	resources := []client.Object{
		&appsv1.DaemonSet{},
		&appsv1.Deployment{},
		&corev1.Service{},
		&storagev1.CSIDriver{},
		&corev1.ConfigMap{},
		&corev1.Secret{},
		&rbacv1.ClusterRoleBinding{},
		&rbacv1.ClusterRole{},
		&corev1.ServiceAccount{},
		&storagev1.StorageClass{},
	}

	for _, obj := range resources {
		if err := a.deleteResourcesByType(ctx, obj, namespace); err != nil {
			log.Error(err, "Failed to delete resource", "type", fmt.Sprintf("%T", obj))
		}
	}

	return nil
}

// createOrUpdate creates or updates a resource
func (a *Actuator) createOrUpdate(ctx context.Context, log logr.Logger, obj client.Object, name string) error {
	log.Info("Creating/Updating resource", "name", name, "type", fmt.Sprintf("%T", obj))

	existing := obj.DeepCopyObject().(client.Object)
	key := client.ObjectKeyFromObject(obj)

	err := a.client.Get(ctx, key, existing)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := a.client.Create(ctx, obj); err != nil {
				return fmt.Errorf("failed to create %s: %w", name, err)
			}
			log.Info("Created resource", "name", name)
			return nil
		}
		return fmt.Errorf("failed to get %s: %w", name, err)
	}

	obj.SetResourceVersion(existing.GetResourceVersion())
	if err := a.client.Update(ctx, obj); err != nil {
		return fmt.Errorf("failed to update %s: %w", name, err)
	}

	log.Info("Updated resource", "name", name)
	return nil
}

// deleteResourcesByType deletes all resources of a given type
func (a *Actuator) deleteResourcesByType(ctx context.Context, obj client.Object, namespace string) error {
	listOpts := []client.DeleteAllOfOption{
		client.InNamespace(namespace),
		client.MatchingLabels{"app.kubernetes.io/name": "synology-csi"},
	}

	list := &corev1.List{}
	list.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	if err := a.client.DeleteAllOf(ctx, obj, listOpts...); err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}
