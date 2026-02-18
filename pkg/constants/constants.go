package constants

const (
	// ExtensionType is the name of the extension type
	ExtensionType = "csi-driver-synology"

	// ExtensionName is the name of the service
	ExtensionName = "gardener-extension-" + ExtensionType

	// GroupName is the name used for gardener group
	GroupName = "csi-driver-synology.metal.extensions.config.gardener.cloud"

	// CSIDriverName is the name of the CSI driver
	CSIDriverName = "csi.san.synology.com"

	// SecretName is the name of the secret containing Synology credentials
	SecretName = "synology-csi-credentials"

	// ClientInfoSecretName is the name of the secret containing client info
	ClientInfoSecretName = "synology-csi-client-info"

	// ConfigMapName is the name of the ConfigMap containing CSI driver config
	ConfigMapName = "synology-csi-config"

	// ChartName is the name of the Helm chart
	ChartName = "csi-driver-synology"

	// ControllerName is the name of the CSI controller
	ControllerName = "synology-csi-controller"

	// NodeName is the name of the CSI node
	NodeName = "synology-csi-node"

	// ProvisionerName is the name of the provisioner
	ProvisionerName = CSIDriverName

	// AttacherName is the name of the attacher
	AttacherName = "synology-csi-attacher"

	// ResizerName is the name of the resizer
	ResizerName = "synology-csi-resizer"

	// SnapshotterName is the name of the snapshotter
	SnapshotterName = "synology-csi-snapshotter"

	// ImageCSIDriver is the image for the Synology CSI driver
	ImageCSIDriver = "synology/synology-csi:v1.1.2"

	// ImageCSIProvisioner is the image for the CSI provisioner
	ImageCSIProvisioner = "registry.k8s.io/sig-storage/csi-provisioner:v5.1.0"

	// ImageCSIAttacher is the image for the CSI attacher
	ImageCSIAttacher = "registry.k8s.io/sig-storage/csi-attacher:v4.7.0"

	// ImageCSIResizer is the image for the CSI resizer
	ImageCSIResizer = "registry.k8s.io/sig-storage/csi-resizer:v1.12.0"

	// ImageCSISnapshotter is the image for the CSI snapshotter
	ImageCSISnapshotter = "registry.k8s.io/sig-storage/csi-snapshotter:v8.1.0"

	// ImageCSINodeDriverRegistrar is the image for the CSI node driver registrar
	ImageCSINodeDriverRegistrar = "registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.12.0"

	// ImageCSILivenessProbe is the image for the CSI liveness probe
	ImageCSILivenessProbe = "registry.k8s.io/sig-storage/livenessprobe:v2.14.0"

	//
	SynologySecretAdminUserRef     = "adminUser"
	SynologySecretAdminPasswordRef = "adminPassword"

	SynologySecretShootUserRef     = "user"
	SynologySecretShootPasswordRef = "password"
)
