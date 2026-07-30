# MWS Cloud Platform Export Example

This example demonstrates how to use the MWS Cloud Platform Packer Export Post-Processor to export images to object storage.

## Overview

The configuration creates a virtual machine image and then exports it to object storage in QCOW2 format.

## Usage

Initialize the Packer plugin:

```bash
packer init .
```

Build and export the image by specifying required variables:

```bash
packer build -var project=YOUR_PROJECT_ID -var service_account_authorized_key_path=/path/to/your/key.dms -var service_account=YOUR_SERVICE_ACCOUNT -var object_storage_bucket=YOUR_BUCKET .
```

Alternatively, you can specify variables in a file:

```bash
packer build -var-file=variables.pkrvars.hcl .
```

The example declares all required variables directly in the configuration file, with sensible defaults for all other parameters.

## Required Variables

- `project` - Your MWS Cloud Platform project ID
- `service_account_authorized_key_path` - Path to your service account key file
- `service_account` - Your service account name
- `object_storage_bucket` - Your object storage bucket for exported image

## Prerequisites

1. A service account with permissions to create resources (provided via service_account_authorized_key_path)
2. A service account with permission to write to object storage (provided via service_account)
3. Object storage bucket.
