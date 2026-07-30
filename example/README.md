# MWS Cloud Platform Packer Plugin Examples

This directory contains example Packer configurations that demonstrate how to use the MWS Cloud Platform Packer Plugin for various operations.

## Prerequisites

Before using these examples, you'll need:

1. Packer installed on your system
2. An MWS Cloud Platform account
3. A service account with appropriate permissions
4. A service account authorized key file

## Usage

Each example is contained in its own directory with a detailed README:

- [Builder Example](./builder/README.md) - Creating VM images
- [Import Example](./import/README.md) - Importing images from object storage
- [Export Example](./export/README.md) - Exporting images to object storage

## Security Notes

- Never commit your service account key files or personal configuration to version control
- The examples are designed to require explicit variable specification rather than hardcoded values
- Always use the principle of least privilege when creating service accounts for Packer operations
