# MWS Cloud Platform Import Example

This example demonstrates how to use the MWS Cloud Platform Packer Import Post-Processor to import images from MWS Cloud Platform Object Storage.

## Overview

The configuration imports an existing image file (QCOW2 format) from MWS Cloud Platform Object Storage into your MWS Cloud Platform project as a new image.

## Usage

Initialize the Packer plugin:

```bash
packer init .
```

Import the image by specifying required variables:

```bash
packer build -var project=YOUR_PROJECT_ID -var service_account_authorized_key_path=/path/to/your/key.dms -var service_account=YOUR_SERVICE_ACCOUNT -var import_object_storage_path=path/to/image.qcow2 .
```

Alternatively, you can specify variables in a file:

```bash
packer build -var-file=variables.pkrvars.hcl .
```

The example declares all required variables directly in the configuration file, with a sensible default for the display name.

## Required Variables

- `project` -  MWS Cloud Platform project ID for creation of virtual machine
- `service_account_authorized_key_path` - Path to authorized key for service account on whose behalf Packer will perform operations in the MWS Cloud
- `service_account` - Name of the service account that will be used to access the MWS Cloud Platform Object Storage
- `import_object_storage_path` - Path in MWS Cloud Platform Object Storage to the image for import (e.g., "bucket-name/path/to/image.qcow2")

## Prerequisites

1. An image file (QCOW2 format) uploaded to MWS Cloud Platform Object Storage
2. A service account with permissions to create resources (provided via `service_account_authorized_key_path`)
3. A service account with permission to read from MWS Cloud Platform Object Storage (provided via `service_account`)
4. The path in MWS Cloud Platform Object Storage where your image is located
