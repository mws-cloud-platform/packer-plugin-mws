Artifact BuilderId: `packer.mws`

The `mws` Packer builder creates [images](https://mws.ru/docs/cloud-platform/compute/general/images-overview.html) for use with [MWS Cloud Platform Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) based on existing images.

## Authentication

Builder supports authentication using:

- Service Account Authorized Key File
- IAM Token
- Instance Metadata Service

### Authentication using Service Account Authorized Key File

To authenticate as a service account, you can set the path to its [authorized
key](https://mws.ru/docs/cloud-platform/iam/keys.html#authkey) using either the
`service_account_authorized_key_path` configuration field or the
`MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` environment variable.

### Authentication using IAM Token

To authenticate using an [IAM
token](https://mws.ru/docs/cloud-platform/iam/sa-get-access-token.html), you can
set the `token` configuration field or the `MWS_TOKEN` environment variable.

### Authentication using Instance Metadata Service

If none of the listed authentication methods is used, plugin will try to detect
if the current environment is a Compute VM with an [attached service
account](https://mws.ru/docs/cloud-platform/compute/general/vm-add-change-delete-sa.html)
by performing a request to the [instance metadata
service](https://mws.ru/docs/cloud-platform/compute/general/vm-metadata-overview.html).
If the request succeeds, plugin will use credentials from the metadata service
for authentication.

## Configuration Reference

Configuration options are organized below into two categories: required and
optional.

In addition to the options listed here, a
[communicator](https://developer.hashicorp.com/packer/docs/communicators) can be
configured for this builder.

<!-- Builder Configuration Fields -->

**Optional**

<!-- Code generated from the comments of the AccessConfig struct in internal/config/access.go; DO NOT EDIT MANUALLY -->

- `project` (string) - The project identifier where resources will be created.
  Can be specified using the `MWS_PROJECT` environment variable.

- `zone` (string) - The zone in which the VM will be created (defaults to "ru-central1-a")

- `base_endpoint` (string) - MWS Cloud Platform API base endpoint (defaults to "https://api.mwsapis.ru").
  Can be specified using the `MWS_BASE_ENDPOINT` environment variable.

- `service_account_authorized_key_path` (string) - Path to the service account authorized key file used for authentication.
  Has no effect if IAM token is set.
  Can be specified using the `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` environment variable.

- `token` (string) - IAM token used for authentication.
  Can be specified using the `MWS_TOKEN` environment variable.

<!-- End of code generated from the comments of the AccessConfig struct in internal/config/access.go; -->

<!-- Code generated from the comments of the DiskConfig struct in internal/config/disk.go; DO NOT EDIT MANUALLY -->

- `disk_name` (string) - Name for the disk (defaults to "packer-{{uuid}}-disk").

- `disk_type` (string) - Type of disk to create (defaults to "nbs-pl2").

- `disk_size` (string) - Size of the disk (defaults to source minDiskSize, in export post-processor minDiskSize of image_for_export is added).

- `disk_iops` (int64) - IOPS for the disk (defaults to 1000).

- `source_project` (string) - Project ID where the source_image/source_disk_backup exists (defaults to the `project`).

- `source_image` (string) - ID of an existing image to use as a base (required unless using `source_disk_backup`).

- `source_disk_backup` (string) - ID of an existing disk backup to use as a base (required unless using `source_image`).

<!-- End of code generated from the comments of the DiskConfig struct in internal/config/disk.go; -->

<!-- Code generated from the comments of the NetworkConfig struct in internal/config/network.go; DO NOT EDIT MANUALLY -->

- `network_name` (string) - Name for the network (defaults to "packer-{{uuid}}-network").
  If specified, Packer will use existing network.

- `subnet_name` (string) - Name for the subnet (defaults to "packer-{{uuid}}-subnet").
  If specified, Packer will use existing subnet.

- `subnet_cidr` (string) - Subnet CIDR (defaults to "192.168.0.0/16").

- `use_external_address` (bool) - Use external address for connection to virtual machine from internet (defaults to "false").

- `external_address_name` (string) - External address name (defaults to "packer-{{uuid}}-external-address").
  Can be specified only if external address usage is enabled.

- `nat64_enable` (bool) - Enables virtual machine ip conversion from ipv4 to ipv6 with RFC 6052 (defaults to "false").
  Meant to be used when packer is in ipv6 only network.

- `nat64_ipv6_prefix` (string) - Prefix used in nat64 conversion (defaults to "64:ff9b::/96" (RFC 6052 Well-Known Prefix)).
  CIDR notation only.

<!-- End of code generated from the comments of the NetworkConfig struct in internal/config/network.go; -->

<!-- Code generated from the comments of the VirtualMachineConfig struct in internal/config/virtual_machine.go; DO NOT EDIT MANUALLY -->

- `virtual_machine_name` (string) - Name for the temporary build VM (defaults to "packer-{{uuid}}-vm").

- `vm_type` (string) - The VM type (defaults to "gen-2-8").

- `cloud_config` (string) - Configuration script for initial setup of a virtual machine in the
  [#cloud-config](https://docs.cloud-init.io/en/latest/explanation/format/cloud-config.html)
  format. Note that this configuration would be extended with SSH key used
  for Packer communicator.

- `vm_service_account` (string) - Service account can be connected to virtual machine so that applications and scripts
  on a virtual machine can work with MWS Cloud Platform services.

- `serial_console_log_file` (string) - File path to save virtual machine serial port output.

- `cleanup_timeout` (duration string | ex: "1h5m2s") - Timeout for resources cleanup (defaults to "1h").

<!-- End of code generated from the comments of the VirtualMachineConfig struct in internal/config/virtual_machine.go; -->

<!-- Code generated from the comments of the ImageConfig struct in internal/config/image.go; DO NOT EDIT MANUALLY -->

- `image_name` (string) - Name for the resulting image (defaults to "packer-{{uuid}}-image").

- `image_display_name` (string) - Display name for the resulting image (defaults to the `image_name`).

- `image_description` (string) - Description for the resulting image. (defaults to "Image created by Packer").

<!-- End of code generated from the comments of the ImageConfig struct in internal/config/image.go; -->


### Example Usage

```hcl
source "mws" "example" {
  project = "your-project"
  service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"

  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20260324"
  
  use_external_address = true
}

build {
  sources = ["source.mws.example"]
}
```
