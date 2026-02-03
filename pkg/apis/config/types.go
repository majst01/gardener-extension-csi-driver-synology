package config

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains configuration for the Synology CSI driver extension
type Configuration struct {
    metav1.TypeMeta
    
    // SynologyHost is the hostname or IP address of the Synology NAS
    SynologyHost string
    
    // SynologyPort is the port of the Synology NAS (default: 5000 for HTTP, 5001 for HTTPS)
    SynologyPort int
    
    // SynologySSL indicates whether to use HTTPS
    SynologySSL bool
    
    // ChapEnabled indicates whether CHAP authentication is enabled
    ChapEnabled bool
    
    // AdminUsername is the admin username for creating shoot-specific users
    AdminUsername string
    
    // AdminPassword is the admin password for creating shoot-specific users
    AdminPassword string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControllerConfiguration defines the configuration for the CSI driver controller
type ControllerConfiguration struct {
    metav1.TypeMeta
    
    // ClientConnection specifies the kubeconfig file and client connection
    // settings for the proxy server to use when communicating with the apiserver.
    // +optional
    ClientConnection *ClientConnection
}

// ClientConnection specifies the kubeconfig file and client connection settings
type ClientConnection struct {
    // Kubeconfig is the path to a kubeconfig file.
    Kubeconfig string
    // QPS controls the number of queries per second allowed
    QPS float32
    // Burst controls the burst for throttle
    Burst int
}
