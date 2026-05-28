# Copyright 2026 MTS Web Services, LLC.
# SPDX-License-Identifier: MPL-2.0

source "mws" "basic-example" {
}

build {
  sources = [
    "source.mws.basic-example"
  ]

  provisioner "shell-local" {
    inline = [
      "echo build generated data: ${build.GeneratedMockData}",
    ]
  }
}
