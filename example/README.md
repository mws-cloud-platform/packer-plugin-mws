# MWS Cloud Platform Builder Example

This directory contains an example Packer configuration that demonstrates how to
use the MWS Cloud Platform Compute Builder. The example includes simple shell
provisioner that echos "Hello!". You can customize provisioners based on your
needs.

## Usage

To use this example:

1. Update the configuration values in `build.pkr.hcl` with your actual values, most important to replace are:
   - `project`
   - `service_account_authorized_key_path`
   - `source_image`
2. Run `packer init .` to install the required plugins
3. Run `packer build .` to build the image
