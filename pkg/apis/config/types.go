package config

import (
	apisconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration defines the configuration for the CSI driver controller
type ControllerConfiguration struct {
	metav1.TypeMeta

	// Synology holds the Synology-specific configuration block ("synology:" im YAML)
	SynologyConfig SynologyConfiguration

	// HealthCheckConfig is the config for the health check controller
	HealthCheckConfig *apisconfigv1alpha1.HealthCheckConfig
}

type SynologyConfiguration struct {
	URL            string
	SecretRef      string
	StorageClasses SynologyStorageClasses
}

type SynologyStorageClasses struct {
	ISCSI ISCSIStorageClass
}

type ISCSIStorageClass struct {
	Parameters map[string]string
}
