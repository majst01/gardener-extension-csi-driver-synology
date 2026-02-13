package v1alpha1

import (
	apisconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains configuration for the Synology CSI driver extension
type ControllerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// SynologyURL is the URL of the Synology NAS
	SynologyURL string `json:"synologyURL,omitempty"`

	// AdminUsername is the admin username for creating shoot-specific users
	AdminUsername string `json:"adminUsername,omitempty"`

	// AdminPassword is the admin password for creating shoot-specific users
	AdminPassword string `json:"adminPassword,omitempty"`

	// HealthCheckConfig is the config for the health check controller
	// +optional
	HealthCheckConfig *apisconfigv1alpha1.HealthCheckConfig `json:"healthCheckConfig,omitempty"`
}
