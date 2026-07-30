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

# Required variables for authentication
variable "project" {
  type = string
}

variable "service_account_authorized_key_path" {
  type = string
}

source "mws" "example" {
  project                             = var.project
  service_account_authorized_key_path = var.service_account_authorized_key_path

  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20260324"

  vm_type = "gen-2-8"

  disk_type = "nbs-pl2"
  disk_size = "10 GB"
  disk_iops = 1000

  use_external_address = true
}

build {
  sources = ["source.mws.example"]

  provisioner "shell" {
    inline = [
      "echo 'Hello from MWS Cloud Platform!'",
    ]
  }
}