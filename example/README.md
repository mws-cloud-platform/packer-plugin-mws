# MWS Cloud Platform Builder Example

This directory contains an example Packer configuration that demonstrates how to use the MWS Cloud Platform Compute Builder.

## Using existing resources

When you specify identifiers for network or subnet resources (`network_name` and `subnet_name`), MWS builder uses those existing resources.

## Configuration Options

### Required Configuration

- `project` - The ID of your MWS project where resources will be created.

### Authentication Configuration

- `service_account_authorized_key_path` - Path to the service account authorized key file used for authentication.
  Has no effect if Token is not empty.
  Can be specified using the `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` environment variable.
- `token` - IAM token for authentication. Can be specified using the `MWS_TOKEN` environment variable.
- `base_endpoint` - MWS Cloud Platform API base endpoint. (defaults to "https://api.mwsapis.ru")
  Can be specified using the `MWS_BASE_ENDPOINT` environment variable.

### Resource References

- `source_image` - ID of an existing image to use as a base (required unless using `source_snapshot`).
- `source_snapshot` - ID of an existing snapshot to use as a base (required unless using `source_image`).

### Auto-created Resources

Names for these resources will be automatically generated if not specified:

- `disk_name` - Name for the disk.
- `external_address_name` - Name for the external IP.
- `network_name` - Name for the network.
- `subnet_name` - Name for the subnet.
- `virtual_machine_name` - Name for the temporary build VM.
- `image_name` - Name for the resulting image.

### Other Optional Configuration

- `zone` - The zone where the VM will be created (defaults to "ru-central1-a").
- `vm_type` - Type of VM to create (defaults to "gen-2-8").
- `disk_type` - Type of disk to create (defaults to "nbs-pl2").
- `disk_size` - Size of the disk (defaults to "10 GB").
- `disk_iops` - IOPS for the disk (defaults to 1000).
- `source_project` - Project ID where the source image/snapshot exists (defaults to the same as `project`).
- `subnet_cidr` - CIDR block for the subnet (defaults to "192.168.0.0/16").
- `image_description` - Description for the resulting image. (defaults to "Image created by Packer").
- `cleanup_timeout` - Timeout for cleanup of create virtual machine step (defaults to "1h").

Note: You must provide exactly one of `source_image` or `source_snapshot`.

## Provisioning

The example includes simple shell provisioner that echos "Hello!". You can customize provisioners based on your needs.

## Usage

To use this example:

1. Update the configuration values in `build.pkr.hcl` with your actual values, most important to replace are:
   - `project`
   - `service_account_authorized_key_path`
   - `source_image`
2. Run `packer init .` to install the required plugins
3. Run `packer build .` to build the image
