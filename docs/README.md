
The MWS Cloud Platform plugin lets you create custom images for use within MWS Cloud Platform Compute.

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    name = {
      source  = "github.com/mws-cloud-platform/mws"
      version = ">=0.0.1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/mws-cloud-platform/mws
```

### Components

#### Builders

- [mws](/packer/integrations/hashicorp/mws/latest/components/builder/mws) - The mws builder creates images from existing ones, by launching an instance, provisioning it, then exporting it as a reusable image.
