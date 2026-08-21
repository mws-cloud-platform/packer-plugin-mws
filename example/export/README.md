# MWS Cloud Platform Export Example

This example demonstrates how to use the MWS Cloud Platform Packer Export Post-Processor to export images created by builder to MWS Cloud Platform Object Storage.

## Overview

The configuration creates a virtual machine image and then exports it to MWS Cloud Platform Object Storage in QCOW2 format.

## Usage

Initialize the Packer plugin:

```bash
packer init .
```

### Using Service Account Authentication (Recommended)

Build and export the image by specifying required variables:

```bash
MWS_PROJECT="YOUR_PROJECT_ID" MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH="/path/to/your/key.dms" packer build -var service_account=YOUR_SERVICE_ACCOUNT -var object_storage_bucket=YOUR_BUCKET .
```

### Using Direct HMAC Key Authentication (Does not require service account with iam admin role)

Build and export the image by specifying HMAC key variables:

```bash
MWS_PROJECT="YOUR_PROJECT_ID" MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH="/path/to/your/key.dms" packer build -var access_key=YOUR_ACCESS_KEY -var secret_key=YOUR_SECRET_KEY -var object_storage_bucket=YOUR_BUCKET .
```

### Alternatively, you can specify variables in a file:

```bash
MWS_PROJECT="YOUR_PROJECT_ID" MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH="/path/to/your/key.dms" packer build -var-file=variables.pkrvars.hcl .
```

The example declares all required variables directly in the configuration file, with sensible defaults for all other parameters.

## Required Environment Variables

- `MWS_PROJECT` - MWS Cloud Platform project ID for creation of virtual machine
- `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` - Path to authorized key for service account on whose behalf Packer will perform operations in the MWS Cloud

## Required Variables

One of the following authentication methods must be provided:

### Option 1: Service Account (Recommended)
- `service_account` - Name of the service account that will be used to access the MWS Cloud Platform Object Storage
- `object_storage_bucket` - Object storage bucket for exported image

### Option 2: Direct HMAC Key Authentication (Does not require service account with iam admin role)
- `access_key` - HMAC key identifier for authenticating with Object Storage
- `secret_key` - HMAC key secret for accessing Object Storage
- `object_storage_bucket` - Object storage bucket for exported image

## Prerequisites

1. A service account with permissions to create resources (provided via `service_account_authorized_key_path`)
2. A service account with permission to write to MWS Cloud Platform Object Storage (provided via `service_account`)
3. Object storage bucket.
