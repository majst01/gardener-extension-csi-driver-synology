package config

import (
	apisconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration defines the configuration for the CSI driver controller
type ControllerConfiguration struct {
	metav1.TypeMeta

	// SynologyURL is the URL of the Synology NAS
	SynologyURL string

	// ChapEnabled indicates whether CHAP authentication is enabled
	ChapEnabled bool

	// AdminUsername is the admin username for creating shoot-specific users
	AdminUsername string

	// AdminPassword is the admin password for creating shoot-specific users
	AdminPassword string

	// HealthCheckConfig is the config for the health check controller
	HealthCheckConfig *apisconfigv1alpha1.HealthCheckConfig
}
