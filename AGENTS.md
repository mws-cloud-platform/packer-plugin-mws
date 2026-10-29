# AGENTS.md

Packer plugin for MWS Cloud Platform Compute Builder.

## Project Structure

```
├── main.go                  # plugin entry point with builder and post-processors registration
├── builder/mws/             # builder (create VM -> provision -> create image)
├── post-processor/
│   ├── mws-export/          # export post-processor (create VM -> attach disk with image -> dump qcow2 -> upload to Object Storage)
│   └── mws-import/          # import post-processor (import image from qcow2 file in Object Storage)
├── internal/                # common packages
├── docs/                    # Documentation sources (rendered via packer-sdc)
├── docs-partials/           # Docs partials (used by `make generate`)
├── example/                 # plugin usage examples
└── .web-docs/               # Web docs
```

## Commands

```shell
make build               # Build the plugin binary (packer-plugin-mws)
make dev                 # Build with `-dev` prerelease and install into the local Packer plugin dir
make install-packer-sdc  # Install packer software development command
make plugin-check        # Validate compiled plugin
make generate            # Generate HCL2 specs, mocks and docs
make test                # Run unit tests
make testacc             # Run all tests (units, acceptance)
```

## Rules

- Generated code (`*.hcl2spec.go`, `*_mock.go`) and documentation (`Code generated from...` blocks) must not be edited by hand — regenerate with `make generate`.
- Config changes require regenerating the HCL2 spec (`make generate`).
- Run unit tests (`make test`) before considering work done.
- Acceptance tests (`make testacc`) requires plugin build (`make dev`) and hit the real MWS Cloud Platform API — never run them by default; only on explicit request.
