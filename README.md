# Gardener Extension for Synology CSI Driver

This extension integrates the [Synology CSI Driver](https://github.com/SynologyOpenSource/synology-csi) with Gardener.

## Features

- Automatic deployment of Synology CSI driver to shoot clusters
- Automatic user creation per shoot cluster on Synology NAS
- iSCSI protocol support with CHAP authentication
- Dynamic volume provisioning
- Volume expansion support
- Volume snapshot support

## Prerequisites

- Synology NAS with DSM 7.0 or later
- iSCSI target service enabled on Synology NAS
- Admin credentials for Synology NAS
- Worker nodes with iSCSI initiator tools installed

### Virtual DSM

can be started with a docker container like so:

```bash
docker run -it --rm --name dsm -e "DISK_SIZE=256G" -p 5000:5000 --device=/dev/kvm --device=/dev/net/tun --cap-add NET_ADMIN -v "${PWD:-.}/dsm:/storage" --stop-timeout 120 docker.io/vdsm/virtual-dsm
```

## Configuration

The extension requires the following configuration in the Garden cluster:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: ControllerDeployment
metadata:
  name: extension-csi-driver-synology
type: helm
providerConfig:
  values:
    synology:
      host: "192.168.1.100"
      port: 5001
      ssl: true
      adminUsername: "admin"
      adminPassword: "your-admin-password"
    chap:
      enabled: true
```

## Usage in Shoot Cluster

After the extension is installed, a default StorageClass synology-iscsi will be available:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: synology-iscsi
```

## Development

Build

```bash
make build
```

Docker image

```bash
make docker-build
make docker-push
```

Install locally

```bash
make install
```

## Licence

Apache License 2.0
