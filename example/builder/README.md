# MWS Cloud Platform Builder Example

This example demonstrates how to use the MWS Cloud Platform Packer Builder to create virtual machine images.

## Overview

The configuration creates a virtual machine from a source image, runs provisioning steps, and saves the result as a new image in your project.

## Usage

Initialize the Packer plugin:

```bash
packer init .
```

Build the image by specifying required variables:

```bash
packer build -var project=YOUR_PROJECT_ID -var service_account_authorized_key_path=/path/to/your/key.dms .
```

Alternatively, you can specify variables in a file:

```bash
packer build -var-file=variables.pkrvars.hcl .
```

The example declares all required variables directly in the configuration file, with sensible defaults for all other parameters.

## Required Variables

- `project` - MWS Cloud Platform project ID for creation of virtual machine
- `service_account_authorized_key_path` - Path to authorized key for service account on whose behalf Packer will perform operations in the MWS Cloud
