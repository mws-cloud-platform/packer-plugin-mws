# MWS Cloud Platform Builder Example

This example demonstrates how to use the MWS Cloud Platform Packer Builder to create virtual machine images.

## Overview

The configuration creates a virtual machine from a source image, runs provisioning steps, and saves the result as a new image in your project.

## Usage

Initialize the Packer plugin:

```bash
packer init .
```

Build the image by specifying required environment variables:

```bash
MWS_PROJECT="YOUR_PROJECT_ID" MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH="/path/to/your/key.dms" packer build .
```

The example declares all required variables directly in the configuration file, with sensible defaults for all other parameters.

## Required Environment Variables

- `MWS_PROJECT` - MWS Cloud Platform project ID for creation of virtual machine
- `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` - Path to authorized key for service account on whose behalf Packer will perform operations in the MWS Cloud
