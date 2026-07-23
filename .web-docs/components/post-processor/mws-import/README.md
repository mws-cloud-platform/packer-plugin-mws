Artifact BuilderId: `packer.post-processor.mws-import`

The `mws-import` Packer post-processor imports images from Object Storage as QCOW2 files to create MWS Cloud Platform images.

## How It Works

The import post-processor follows a sequential process to import images from Object Storage:

1. **HMAC Key Management**: Either creates a temporary HMAC key for Object Storage authentication using the provided service account, or uses the provided access key and secret key credentials.

2. **AWS Client Creation**: Creates an AWS S3 client with the appropriate credentials to access Object Storage.

3. **Signed URL Generation**: Generates a presigned URL for the Object Storage object to allow the MWS Cloud Platform to access the image file.

4. **Image Import**: Imports the image from the Object Storage location using the presigned URL.

5. **Cleanup**: Removes temporary resources including, if applicable, the temporary HMAC key.

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

For importing the image from Object Storage, the post-processor supports two authentication methods:

### Authentication using Service Account (Recommended)

To authenticate with Object Storage using a service account, you can set the `service_account` configuration field. The post-processor will automatically generate a temporary HMAC key for accessing Object Storage.

### Authentication using HMAC Keys

To authenticate with Object Storage using HMAC keys, you can set both the `access_key` and `secret_key` configuration fields.

## Configuration Reference

Configuration options are organized below into two categories: required and
optional.

<!-- Post-Processor Configuration Fields -->

**Required**

<!-- Code generated from the comments of the AccessConfig struct in internal/config/config.go; DO NOT EDIT MANUALLY -->

- `project` (string) - The project identifier where resources will be created.

<!-- End of code generated from the comments of the AccessConfig struct in internal/config/config.go; -->

<!-- Code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-import/config.go; DO NOT EDIT MANUALLY -->

- `object_storage_path` (string) - MWS Cloud Platform Object Storage path from where the image will be imported.

<!-- End of code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-import/config.go; -->


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

<!-- Code generated from the comments of the ImageConfig struct in internal/config/config.go; DO NOT EDIT MANUALLY -->

- `image_name` (string) - Name for the resulting image (defaults to "packer-{{uuid}}-image").

- `image_display_name` (string) - Display name for the resulting image (defaults to the `image_name`).

- `image_description` (string) - Description for the resulting image. (defaults to "Image created by Packer").

<!-- End of code generated from the comments of the ImageConfig struct in internal/config/config.go; -->

<!-- Code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-import/config.go; DO NOT EDIT MANUALLY -->

- `service_account` (string) - MWS Cloud Platform Service Account used for generating temporal HMAC key
  to access Object Storage. Required, unless `access_key` and `secret_key`
  are provided.

- `access_key` (string) - HMAC key identifier for authenticating with Object Storage. Used if
  `service_account` is not provided. Also requires `secret_key` to be
  provided.

- `secret_key` (string) - HMAC key secret for accessing Object Storage. Required if `access_key` is
  provided.

- `object_storage_endpoint` (string) - MWS Cloud Platform Object Storage endpoint to import image from (defaults to "https://storage.mwsapis.ru").

- `object_storage_region` (string) - MWS Cloud Platform Object Storage region where the bucket is located (defaults to "ru-central1").

<!-- End of code generated from the comments of the ObjectStorageConfig struct in post-processor/mws-import/config.go; -->

<!-- Code generated from the comments of the Config struct in post-processor/mws-import/config.go; DO NOT EDIT MANUALLY -->

- `cleanup_timeout` (duration string | ex: "1h5m2s") - Timeout for resources cleanup (defaults to "1h").

<!-- End of code generated from the comments of the Config struct in post-processor/mws-import/config.go; -->


### Example Usage

#### Using Service Account for Object Storage Authentication (Recommended)

```hcl
source "null" "empty" {
  communicator = "none"
}

build {
  sources = ["source.null.empty"]

  post-processor "mws-import" {
    project = "your-project"
    service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"
    
    service_account = "your-service-account"
    object_storage_path = "your-bucket/image.qcow2"
    object_storage_endpoint = "https://storage.mwsapis.ru"
    
    image_display_name = "Imported Image"
  }
}
```

#### Using HMAC Keys for Object Storage Authentication

```hcl
source "null" "empty" {
  communicator = "none"
}

build {
  sources = ["source.null.empty"]

  post-processor "mws-import" {
    project = "your-project"
    service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"
    
    access_key = "your-access-key"
    secret_key = "your-secret-key"
    object_storage_path = "your-bucket/image.qcow2"
    object_storage_endpoint = "https://storage.mwsapis.ru"
    
    image_display_name = "Imported Image"
  }
}
