# Copyright 2026 MTS Web Services, LLC.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    mws = {
      version = ">= 0.5.0"
      source  = "github.com/mws-cloud-platform/mws"
    }
  }
}

variable "project" {
  type = string
}

variable "service_account_authorized_key_path" {
  type = string
}

# Variables specific to export operations
variable "service_account" {
  type = string
}

variable "object_storage_bucket" {
  type = string
}

source "mws" "example" {
  project                             = var.project
  service_account_authorized_key_path = var.service_account_authorized_key_path

  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20260324"

  use_external_address = true
}

build {
  sources = ["source.mws.example"]

  provisioner "shell" {
    inline = [
      "echo 'Preparing image for export...'",
    ]
  }

  post-processor "mws-export" {
    project                             = var.project
    service_account_authorized_key_path = var.service_account_authorized_key_path

    source_project = "mws-ubuntu"
    source_image   = "mws-ubuntu-2404-lts-v20260324"
    disk_size      = "20 GB"

    use_external_address = true

    service_account     = var.service_account
    object_storage_path = "${var.object_storage_bucket}/{{build `ImageName` }}.qcow2"
  }
}
