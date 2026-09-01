# Copyright 2026 MTS Web Services, LLC.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    mws = {
      version = ">= 0.8.0"
      source  = "github.com/mws-cloud-platform/mws"
    }
  }
}

variable "image_name" {
  type    = string
  default = ""
}

# Variables specific to import operations
variable "import_object_storage_path" {
  type = string
}

variable "service_account" {
  type    = string
  default = ""
}

variable "access_key" {
  type    = string
  default = ""
}

variable "secret_key" {
  type    = string
  default = ""
}

source "null" "empty" {
  communicator = "none"
}

build {
  sources = ["source.null.empty"]

  post-processor "mws-import" {
    service_account     = var.service_account
    access_key          = var.access_key
    secret_key          = var.secret_key
    object_storage_path = var.import_object_storage_path

    image_name         = var.image_name
    image_display_name = "Imported image from object storage"
  }
}