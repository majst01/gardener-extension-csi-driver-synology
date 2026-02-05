package v1alpha1

import (
	apisconfigv1alpha1 "github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ShootConfiguration defines the configuration for the CSI driver in the shoot cluster
type ShootConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// SynologyURL is the URL of the Synology NAS
	SynologyURL string `json:"synologyUrl,omitempty"`

	// ChapEnabled indicates whether CHAP authentication is enabled
	ChapEnabled bool `json:"chapEnabled,omitempty"`

	// Username is the username for creating shoot-specific volumes
	Username string `json:"username,omitempty"`

	// Password is the password for creating shoot-specific volumes
	Password string `json:"password,omitempty"`

	// HealthCheckConfig is the config for the health check controller
	// +optional
	HealthCheckConfig *apisconfigv1alpha1.HealthCheckConfig `json:"healthCheckConfig,omitempty"`
}
