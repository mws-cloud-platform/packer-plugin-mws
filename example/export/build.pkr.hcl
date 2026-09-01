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

# Variables specific to export operations
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

variable "object_storage_bucket" {
  type = string
}

source "mws" "example" {
  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20260324"

  use_external_address = true

  image_name = var.image_name
}

build {
  sources = ["source.mws.example"]

  provisioner "shell" {
    inline = [
      "echo 'Preparing image for export...'",
    ]
  }

  post-processor "mws-export" {
    source_project = "mws-ubuntu"
    source_image   = "mws-ubuntu-2404-lts-v20260324"

    use_external_address = true

    service_account     = var.service_account
    access_key          = var.access_key
    secret_key          = var.secret_key
    object_storage_path = "${var.object_storage_bucket}/{{build `ImageName` }}.qcow2"
  }
}
