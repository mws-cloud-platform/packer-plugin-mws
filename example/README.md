# MWS Cloud Platform Packer Plugin Examples

This directory contains example Packer configurations that demonstrate how to use the MWS Cloud Platform Packer Plugin for various operations.

## Prerequisites

Before using these examples, you'll need:

<!-- TODO: replace first link with our public doc -->
1. [Packer installed on your system](https://developer.hashicorp.com/packer/install)
2. [An MWS Cloud Platform account](https://mws.ru/docs/cloud-platform/about/quickstart.html)
3. [A service account with appropriate permissions](https://mws.ru/docs/cloud-platform/iam/sa.html)
4. [A service account authorized key file](https://mws.ru/docs/cloud-platform/iam/keys.html#authkey)

## Usage

Each example is contained in its own directory with a detailed README:

- [Builder Example](./builder/README.md) - Creating VM images
- [Import Example](./import/README.md) - Importing images from MWS Cloud Platform Object Storage
- [Export Example](./export/README.md) - Exporting images created by builder to MWS Cloud Platform Object Storage
- [Export Example](./export-of-existing-image/README.md) - Exporting existing images to MWS Cloud Platform Object Storage

## Security Notes

- Never commit your service account key files or personal configuration to version control
- The examples are designed to require explicit variable specification rather than hardcoded values
- Always use the principle of least privilege when creating service accounts for Packer operations
