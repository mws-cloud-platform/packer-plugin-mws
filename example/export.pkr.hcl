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
  project      = "your-project-id"
  source_image = "your-source-image-id"
  zone         = "ru-central1-a"

  # Additional configuration as needed
  vm_type      = "gen-2-8"
  disk_type    = "nbs-pl2"
  disk_size    = "10 GB"
}

build {
  sources = ["source.mws.example"]

  provisioner "shell" {
    inline = [
      "echo 'Hello'",
    ]
  }

  post-processor "mws-export" {
    service_account = "your-service-account"
    s3_bucket       = "your-s3-bucket-name"
  }
}