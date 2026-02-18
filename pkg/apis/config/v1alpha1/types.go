package v1alpha1

import (
	apisconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains configuration for the Synology CSI driver extension
type ControllerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// Synology holds the Synology-specific configuration
	SynologyConfig SynologyConfiguration `json:"synology"`

	// HealthCheckConfig is the config for the health check controller
	// +optional
	HealthCheckConfig *apisconfigv1alpha1.HealthCheckConfig `json:"healthCheckConfig,omitempty"`
}

type SynologyConfiguration struct {
	URL       string `json:"url"`
	SecretRef string `json:"secretRef"`

	// StorageClasses defines storage class configuration
	StorageClasses SynologyStorageClasses `json:"storageClasses"`
}

type SynologyStorageClasses struct {
	ISCSI ISCSIStorageClass `json:"iscsi"`
}

type ISCSIStorageClass struct {
	Parameters map[string]string `json:"parameters"`
}
