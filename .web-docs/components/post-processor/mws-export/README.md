Artifact BuilderId: `packer.post-processor.mws-export`

The `mws-export` Packer post-processor exports images created by the MWS Cloud Platform builder to Object Storage as QCOW2 files.

## How It Works

The export post-processor follows a sequential process to export images to Object Storage:

1. **SSH Key Generation**: Generates temporary SSH keys for secure communication with the virtual machine.

2. **HMAC Key Management**: Either creates a temporary HMAC key for Object Storage authentication using the provided service account, or uses the provided access key and secret key credentials.

3. **Virtual Machine Creation**: Creates a temporary virtual machine with the specified configuration.

4. **Disk Attachment**: Creates a disk from the source image and attaches it to the virtual machine.

5. **SSH Connection**: Establishes an SSH connection to the virtual machine.

6. **Tool Installation**: Installs the required tools (`qemu-img` and AWS CLI) if they're not already present on the virtual machine.

7. **Image Dumping**: Dumps the attached disk image to a QCOW2 file using `qemu-img`.

8. **File Upload**: Uploads the QCOW2 file to the specified Object Storage location using the AWS CLI with the configured credentials based on the provided HMAC key.

9. **Cleanup**: Removes temporary resources including SSH keys and, if applicable, the temporary HMAC key.

## Authentication

The post-processor supports authentication using:

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

If none of the listed authentication methods is used, post-processor will try to detect
if the current environment is a Compute VM with an [attached service
account](https://mws.ru/docs/cloud-platform/compute/general/vm-add-change-delete-sa.html)
by performing a request to the [instance metadata
service](https://mws.ru/docs/cloud-platform/compute/general/vm-metadata-overview.html).
If the request succeeds, post-processor will use credentials from the metadata service
for authentication.

## Object Storage Authentication

For uploading the exported image to Object Storage, the post-processor supports two authentication methods:

### Authentication using Service Account (Recommended)

To authenticate with Object Storage using a service account, you can set the `service_account` configuration field. The post-processor will automatically generate a temporary HMAC key for accessing Object Storage.

### Authentication using HMAC Keys

To authenticate with Object Storage using HMAC keys, you can set both the `access_key` and `secret_key` configuration fields.

## Configuration Reference

Configuration options are organized below into two categories: required and
optional.

In addition to the options listed here, a
[communicator](https://developer.hashicorp.com/packer/docs/communicators) can be
configured for this post-processor.

<!-- Post-Processor Configuration Fields -->

**Required**

<!-- Code generated from the comments of the AccessConfig struct in internal/config/config.go; DO NOT EDIT MANUALLY -->

- `project` (string) - The project identifier where resources will be created.

<!-- End of code generated from the comments of the AccessConfig struct in internal/config/config.go; -->

<!-- Code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-export/config.go; DO NOT EDIT MANUALLY -->

- `object_storage_path` (string) - MWS Cloud Platform Object Storage path where the image will be stored.

<!-- End of code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-export/config.go; -->


**Optional**

<!-- Code generated from the comments of the AccessConfig struct in internal/config/config.go; DO NOT EDIT MANUALLY -->

- `zone` (string) - The zone in which the VM will be created (defaults to "ru-central1-a")

- `base_endpoint` (string) - MWS Cloud Platform API base endpoint (defaults to "https://api.mwsapis.ru").
  Can be specified using the `MWS_BASE_ENDPOINT` environment variable.

- `service_account_authorized_key_path` (string) - Path to the service account authorized key file used for authentication.
  Has no effect if IAM token is set.
  Can be specified using the `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` environment variable.

- `token` (string) - IAM token used for authentication.
  Can be specified using the `MWS_TOKEN` environment variable.

<!-- End of code generated from the comments of the AccessConfig struct in internal/config/config.go; -->

<!-- Code generated from the comments of the VirtualMachineConfig struct in internal/config/config.go; DO NOT EDIT MANUALLY -->

- `virtual_machine_name` (string) - Name for the temporary build VM (defaults to "packer-{{uuid}}-vm").

- `vm_type` (string) - The VM type (defaults to "gen-2-8").

- `cloud_config` (string) - Configuration script for initial setup of a virtual machine in the
  [#cloud-config](https://docs.cloud-init.io/en/latest/explanation/format/cloud-config.html)
  format. Note that this configuration would be extended with SSH key used
  for Packer communicator.

- `cleanup_timeout` (duration string | ex: "1h5m2s") - Timeout for resources cleanup (defaults to "1h").

<!-- End of code generated from the comments of the VirtualMachineConfig struct in internal/config/config.go; -->

<!-- Code generated from the comments of the DiskForExportConfig struct in post-processor/mws-export/config.go; DO NOT EDIT MANUALLY -->

- `disk_for_export_type` (string) - Type of the disk used for image export (defaults to "nbs-pl2").

- `disk_for_export_iops` (int64) - IOPS for the disk used for image export (defaults to 1000).

- `image_for_export_project` (string) - The project identifier where the image for export exists (defaults to the `project`).

- `image_for_export` (string) - Identifier of the image to export. Required only when post processor used
  without mws builder.

<!-- End of code generated from the comments of the DiskForExportConfig struct in post-processor/mws-export/config.go; -->

<!-- Code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-export/config.go; DO NOT EDIT MANUALLY -->

- `service_account` (string) - MWS Cloud Platform Service Account used for generating temporal HMAC key
  to access Object Storage. Required, unless `access_key` and `secret_key`
  are provided.

- `access_key` (string) - HMAC key identifier for authenticating with Object Storage. Used if
  `service_account` is not provided. Also requires `secret_key` to be
  provided.

- `secret_key` (string) - HMAC key secret for accessing Object Storage. Required if `access_key` is
  provided.

- `object_storage_endpoint` (string) - MWS Cloud Platform Object Storage endpoint to upload image (defaults to "https://storage.mwsapis.ru").

- `object_storage_region` (string) - MWS Cloud Platform Object Storage region where the bucket is located (defaults to "ru-central1").

<!-- End of code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-export/config.go; -->


### Example Usage

#### Using Service Account for Object Storage Authentication (Recommended)

```hcl
source "mws" "example" {
  project = "your-project"
  service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"

  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20250529"

  use_external_address = true
}

build {
  sources = ["source.mws.example"]

  post-processor "mws-export" {
    project = "your-project"
    service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"
    
    source_project = "mws-ubuntu"
    source_image   = "mws-ubuntu-2404-lts-v20260526"
    disk_size      = "20 GB"
    use_external_address = true

    service_account = "your-service-account"
    object_storage_path = "your-bucket/{{build `ImageName` }}.qcow2"
    object_storage_endpoint = "https://storage.mwsapis.ru"
  }
}
```

#### Using HMAC Keys for Object Storage Authentication

```hcl
source "mws" "example" {
  project = "your-project"
  service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"

  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20250529"
  
  use_external_address = true
}

build {
  sources = ["source.mws.example"]

  post-processor "mws-export" {
    project = "your-project"
    service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"
    
    source_project = "mws-ubuntu"
    source_image   = "mws-ubuntu-2404-lts-v20260526"
    disk_size      = "20 GB"
    use_external_address = true
    
    access_key = "your-access-key"
    secret_key = "your-secret-key"
    object_storage_path = "your-bucket/{{build `ImageName` }}.qcow2"
    object_storage_endpoint = "https://storage.mwsapis.ru"
  }
}
```
