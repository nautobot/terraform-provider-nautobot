# Terraform Provider Nautobot

## Requirements

* [Terraform](https://www.terraform.io/downloads.html) or [OpenTofu](https://opentofu.org/docs/intro/install/). The versions used for acceptance tests are defined in the [Tests workflow](.github/workflows/test.yml).
* [Go](https://golang.org/doc/install). The required Go version is defined in [`go.mod`](go.mod).

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the `make` command:

```sh
$ make install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up-to-date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

The provide requires two arguments, `url` and `token`. For the data sources and resources supported, take a look at the [internal/provider](internal/provider) folder. In the next example, we capture the data of all manufacturers and create a new manufacturer "Vendor I". For all arguments that the provider accepts, see its [documentation](docs/index.md)

```hcl
terraform {
  required_providers {
    nautobot = {
      version = "3.1.0"
      source  = "registry.terraform.io/nautobot/nautobot"
    }
  }
}

provider "nautobot" {
  url = "https://demo.nautobot.com/api/"
  token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}

data "nautobot_manufacturers" "all" {}

resource "nautobot_manufacturer" "new" {
  description = "Created with Terraform"
  name    = "Vendor I"
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

Please read the [contributor guide](CONTRIBUTORS.md) before opening a pull request.

There are a few make targets you can leverage:

- `make install`: To compile the provider.
- `make fmt`: Format all tracked Go files with `gofmt`.
- `make fmt-check`: Check Go formatting without modifying files.
- `make docs`: To generate or update documentation.
- `make local`: Test local version of the provider.
- `make testacc`: Start or reuse a local Nautobot instance and run the full acceptance test suite against it.
- `make testacc-run`: Run acceptance tests against an already-running Nautobot instance without managing Compose.
- `make testacc-local-up`: Start the reusable local Nautobot instance without running tests.
- `make testacc-local-down`: Remove the local Nautobot containers and volumes.

_Note:_ Acceptance tests create real objects in the local Nautobot instance. An interrupted or
failed run may leave test objects behind until the instance is removed.

```sh
$ make testacc
```

OpenTofu is used when available; otherwise Terraform is used. Select one explicitly with
`TF_TOOL=opentofu make testacc` or `TF_TOOL=terraform make testacc`.

To rerun tests without invoking Compose, use:

```sh
$ make testacc-run
```

Individual acceptance tests can be selected with `TEST` and `TESTARGS`:

```sh
$ make testacc-run TEST=./internal/provider/... TESTARGS='-run TestAccTenantResource_drift'
```

By default, `testacc-run` connects to `http://localhost:8080`. Override
`NAUTOBOT_TEST_URL` and `NAUTOBOT_TEST_TOKEN` when targeting another instance.

The local Nautobot instance remains running after the tests finish or fail, so subsequent
acceptance-test runs reuse it. Remove it explicitly with:

```sh
$ make testacc-local-down
```

## Credits

This [project](https://github.com/nleiva/terraform-provider-nautobot) started as an exercise for educational purposes by @nleiva during the development of his book "Network Automation with Go". Thank you Nicolas for your effort and collaboration!
