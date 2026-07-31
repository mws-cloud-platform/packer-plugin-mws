# Copyright 2026 MTS Web Services, LLC.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    mws = {
      version = ">= 0.6.0"
      source  = "github.com/mws-cloud-platform/mws"
    }
  }
}

# Required variables for authentication
variable "project" {
  type = string
}

variable "service_account_authorized_key_path" {
  type = string
}

# Variables specific to import operations
variable "import_object_storage_path" {
  type = string
}

variable "service_account" {
  type = string
}

source "null" "empty" {
  communicator = "none"
}

build {
  sources = ["source.null.empty"]

  post-processor "mws-import" {
    project                             = var.project
    service_account_authorized_key_path = var.service_account_authorized_key_path

    service_account     = var.service_account
    object_storage_path = var.import_object_storage_path

    image_display_name = "Imported image from object storage"
  }
}