Artifact BuilderId: `packer.mws`

<!--
  Include a short description about the builder. This is a good place
  to call out what the builder does, and any requirements for the given
  builder environment. See https://www.packer.io/docs/builder/null
-->

The `mws` Packer builder is able to create [images](https://mws.ru/docs/cloud-platform/compute/general/images-overview.html) for use with [MWS Cloud Platform Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) based on existing images.

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

If none of the listed authentication methods is used, builder will try to detect
if the current environment is a Compute VM with an [attached service
account](https://mws.ru/docs/cloud-platform/compute/general/vm-add-change-delete-sa.html)
by performing a request to the [instance metadata
service](https://mws.ru/docs/cloud-platform/compute/general/vm-metadata-overview.html).
If the request succeeds, builder will use credentials from the metadata service
for authentication.

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

In addition to the options listed here, a
[communicator](https://developer.hashicorp.com/packer/docs/communicators) can be
configured for this builder.

<!-- Builder Configuration Fields -->

**Required**

<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

<!-- Code generated from the comments of the Config struct in builder/mws/config.go; DO NOT EDIT MANUALLY -->

- `project` (string) - The project identifier where resources will be created.

<!-- End of code generated from the comments of the Config struct in builder/mws/config.go; -->


**Optional**

<!-- Code generated from the comments of the Config struct in builder/mws/config.go; DO NOT EDIT MANUALLY -->

- `zone` (string) - The zone in which the VM will be created (defaults to "ru-central1-a")

- `base_endpoint` (string) - MWS Cloud Platform API base endpoint (defaults to "https://api.mwsapis.ru").
  Can be specified using the `MWS_BASE_ENDPOINT` environment variable.

- `service_account_authorized_key_path` (string) - Path to the service account authorized key file used for authentication.
  Has no effect if IAM token is set.
  Can be specified using the `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` environment variable.

- `token` (string) - IAM token used for authentication.
  Can be specified using the `MWS_TOKEN` environment variable.

- `virtual_machine_name` (string) - Name for the temporary build VM (defaults to "packer-{{uuid}}-vm").

- `vm_type` (string) - The VM type (defaults to "gen-2-8").

- `image_name` (string) - Name for the resulting image (defaults to "packer-{{uuid}}-image").

- `image_description` (string) - Description for the resulting image. (defaults to "Image created by Packer").

- `disk_name` (string) - Name for the disk (defaults to "packer-{{uuid}}-disk").

- `disk_type` (string) - Type of disk to create (defaults to "nbs-pl2").

- `disk_size` (string) - Size of the disk (defaults to "10 GB").

- `disk_iops` (int64) - IOPS for the disk (defaults to 1000).

- `source_project` (string) - Project ID where the source image/snapshot exists (defaults to the `project`).

- `source_image` (string) - ID of an existing image to use as a base (required unless using `source_snapshot`).

- `source_snapshot` (string) - ID of an existing snapshot to use as a base (required unless using `source_image`).

- `network_name` (string) - Name for the network (defaults to "packer-{{uuid}}-network").
  If specified, Packer will use existing network.

- `subnet_name` (string) - Name for the subnet (defaults to "packer-{{uuid}}-subnet").
  If specified, Packer will use existing subnet.

- `subnet_cidr` (string) - Subnet CIDR (defaults to "192.168.0.0/16").

- `use_external_address` (bool) - Use external address for connection to virtual machine from internet (defaults to "false").

- `external_address_name` (string) - External address name (defaults to "packer-{{uuid}}-external-address").
  Can be specified only if external address usage is enabled.

- `cleanup_timeout` (string) - Timeout for cleanup of create virtual machine step (defaults to "1h").

<!-- End of code generated from the comments of the Config struct in builder/mws/config.go; -->


<!--
  A basic example on the usage of the builder. Multiple examples
  can be provided to highlight various build configurations.

-->

### Example Usage

```hcl
source "mws" "example" {
  project = "your-project"

  service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"

  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20250529"
}

build {
  sources = ["source.mws.example"]
}
```
