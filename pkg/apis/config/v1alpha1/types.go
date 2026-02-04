package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains configuration for the Synology CSI driver extension
type ControllerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// SynologyHost is the hostname or IP address of the Synology NAS
	SynologyHost string `json:"synologyHost"`

	// SynologyPort is the port of the Synology NAS (default: 5000 for HTTP, 5001 for HTTPS)
	SynologyPort int `json:"synologyPort"`

	// SynologySSL indicates whether to use HTTPS
	SynologySSL bool `json:"synologySSL"`

	// ChapEnabled indicates whether CHAP authentication is enabled
	ChapEnabled bool `json:"chapEnabled"`

	// AdminUsername is the admin username for creating shoot-specific users
	AdminUsername string `json:"adminUsername"`

	// AdminPassword is the admin password for creating shoot-specific users
	AdminPassword string `json:"adminPassword"`
}
