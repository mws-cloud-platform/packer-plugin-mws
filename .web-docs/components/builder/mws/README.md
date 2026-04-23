Type: `mws`
Artifact BuilderId: `packer.mws`

<!--
  Include a short description about the builder. This is a good place
  to call out what the builder does, and any requirements for the given
  builder environment. See https://www.packer.io/docs/builder/null
-->

The `mws` Packer builder is able to create [images](https://mws.ru/docs/cloud-platform/compute/general/images-overview.html) for use with [MWS Cloud Platform Compute](https://mws.ru/docs/cloud-platform/compute/general/whatis-compute.html) based on existing images.


<!-- Builder Configuration Fields -->

**Required**


<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

**Optional**


<!--
  A basic example on the usage of the builder. Multiple examples
  can be provided to highlight various build configurations.

-->
### Example Usage


```hcl
 source "mws" "example" {
 }

 build {
   sources = ["source.mws.example"]
 }
```
