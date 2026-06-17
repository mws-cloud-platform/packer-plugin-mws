# Copyright 2026 MTS Web Services, LLC.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    mws = {
      version = ">= 0.1.0"
      source  = "github.com/mws-cloud-platform/mws"
    }
  }
}

source "mws" "example" {
  project = "your-project"
  zone    = "ru-central1-a"

  service_account_authorized_key_path = "/path/to/your/service_account_authorized_key.dms"

  vm_type = "gen-2-8"

  disk_type = "nbs-pl2"
  disk_size = "10 GB"
  disk_iops = 1000

  source_project = "mws-ubuntu"
  source_image   = "mws-ubuntu-2404-lts-v20250529"

  use_external_address = true
}

build {
  sources = ["source.mws.example"]

  provisioner "shell" {
    inline = [
      "echo 'Hello!'",
    ]
  }
}
