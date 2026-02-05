package synology

import (
	"fmt"
	"net/url"

	"github.com/metal-stack/gardener-extension-csi-driver-synology/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

// ManifestConfig contains configuration for generating manifests
type ManifestConfig struct {
	Namespace    string
	Url          string
	Username     string
	Password     string
	ChapEnabled  bool
	ChapUsername string
	ChapPassword string
}

// GenerateNamespace generates the namespace for the CSI driver
func GenerateNamespace(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "synology-csi",
			},
		},
	}
}

// GenerateServiceAccount generates the service account
func GenerateServiceAccount(namespace, name string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "synology-csi",
				"app.kubernetes.io/component": name,
			},
		},
	}
}

// GenerateControllerClusterRole generates the cluster role for the controller
func GenerateControllerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.ControllerName,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "synology-csi",
				"app.kubernetes.io/component": "controller",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "create", "delete", "patch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"csinodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments/status"},
				Verbs:     []string{"patch"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshots"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotcontents"},
				Verbs:     []string{"get", "list", "watch", "update", "patch", "create", "delete"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"snapshot.storage.k8s.io"},
				Resources: []string{"volumesnapshotcontents/status"},
				Verbs:     []string{"update", "patch"},
			},
		},
	}
}

// GenerateNodeClusterRole generates the cluster role for the node
func GenerateNodeClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.NodeName,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "synology-csi",
				"app.kubernetes.io/component": "node",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
		},
	}
}

// GenerateClusterRoleBinding generates the cluster role binding
func GenerateClusterRoleBinding(name, namespace, serviceAccount string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "synology-csi",
				"app.kubernetes.io/component": name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccount,
				Namespace: namespace,
			},
		},
	}
}

// GenerateSecret generates the secret containing Synology credentials
func GenerateSecret(config *ManifestConfig) (*corev1.Secret, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Synology URL: %w", err)
	}

	data := map[string][]byte{
		"host":     []byte(u.Hostname()),
		"port":     []byte(u.Port()),
		"protocol": []byte(u.Scheme),
		"username": []byte(config.Username),
		"password": []byte(config.Password),
	}

	if config.ChapEnabled {
		data["chap-enabled"] = []byte("true")
		data["chap-username"] = []byte(config.ChapUsername)
		data["chap-password"] = []byte(config.ChapPassword)
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.SecretName,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "synology-csi",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}, nil
}

// GenerateConfigMap generates the ConfigMap
func GenerateConfigMap(config *ManifestConfig) (*corev1.ConfigMap, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Synology URL: %w", err)
	}

	clientInfoYAML := fmt.Sprintf(`---
clients:
  - host: %s://%s
    port: %s
    https: %t
`, u.Scheme, u.Hostname(), u.Port(), u.Scheme == "https")

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ConfigMapName,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": "synology-csi",
			},
		},
		Data: map[string]string{
			"client-info.yaml": clientInfoYAML,
		},
	}, nil
}

// GenerateCSIDriver generates the CSIDriver resource
func GenerateCSIDriver() *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.CSIDriverName,
			Labels: map[string]string{
				"app.kubernetes.io/name": "synology-csi",
			},
		},
		Spec: storagev1.CSIDriverSpec{
			AttachRequired: ptr.To(true),
			PodInfoOnMount: ptr.To(false),
			VolumeLifecycleModes: []storagev1.VolumeLifecycleMode{
				storagev1.VolumeLifecyclePersistent,
			},
		},
	}
}

// GenerateStorageClass generates the default StorageClass
func GenerateStorageClass(namespace string) *storagev1.StorageClass {
	reclaimPolicy := corev1.PersistentVolumeReclaimDelete
	volumeBindingMode := storagev1.VolumeBindingImmediate
	allowVolumeExpansion := true

	return &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "synology-iscsi",
			Labels: map[string]string{
				"app.kubernetes.io/name": "synology-csi",
			},
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
		Provisioner:          constants.CSIDriverName,
		ReclaimPolicy:        &reclaimPolicy,
		VolumeBindingMode:    &volumeBindingMode,
		AllowVolumeExpansion: &allowVolumeExpansion,
		Parameters: map[string]string{
			"protocol": "iscsi",
			"fsType":   "ext4",
			"csi.storage.k8s.io/provisioner-secret-name":             constants.SecretName,
			"csi.storage.k8s.io/provisioner-secret-namespace":        namespace,
			"csi.storage.k8s.io/controller-publish-secret-name":      constants.SecretName,
			"csi.storage.k8s.io/controller-publish-secret-namespace": namespace,
			"csi.storage.k8s.io/node-stage-secret-name":              constants.SecretName,
			"csi.storage.k8s.io/node-stage-secret-namespace":         namespace,
			"csi.storage.k8s.io/node-publish-secret-name":            constants.SecretName,
			"csi.storage.k8s.io/node-publish-secret-namespace":       namespace,
		},
	}
}

// GenerateService generates a service for the CSI controller
func GenerateService(namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.ControllerName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "synology-csi",
				"app.kubernetes.io/component": "controller",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name":      "synology-csi",
				"app.kubernetes.io/component": "controller",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "healthz",
					Protocol:   corev1.ProtocolTCP,
					Port:       9808,
					TargetPort: intstr.FromString("healthz"),
				},
			},
		},
	}
}
