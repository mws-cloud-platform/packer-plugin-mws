# Copyright IBM Corp. 2020, 2025
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    mws = {
      version = ">=v0.1.0"
      source  = "github.com/mws-cloud-platform/mws"
    }
  }
}

source "mws" "example" {
}

build {
  sources = ["source.mws.example"]
  provisioner "shell" {
    inline = [
      "echo Hello!"
    ]
  }
}
