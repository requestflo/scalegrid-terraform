# Terraform Provider for ScaleGrid

A Terraform provider for [ScaleGrid](https://scalegrid.io) — manage MongoDB,
Redis, MySQL, and PostgreSQL database deployments (and their cloud profiles,
firewall rules, and backups) as code through the
[ScaleGrid REST API](https://scalegrid.io/api/).

Built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
(protocol 6).

## Features

**Resources**

| Resource | Description |
|----------|-------------|
| `scalegrid_cluster` | A database deployment (MongoDB / Redis / MySQL / PostgreSQL). Provisioning is awaited; scaling and disk growth are applied in place. |
| `scalegrid_cloud_profile` | Cloud credentials for Bring Your Own Cloud deployments (AWS, Azure, GCP, DigitalOcean, OCI, VMware). |
| `scalegrid_firewall_rule` | A CIDR allow rule attached to a cluster. |
| `scalegrid_backup` | An on-demand backup of a cluster. |

**Data sources**

| Data source | Description |
|-------------|-------------|
| `scalegrid_cluster` | Look up a single cluster by ID. |
| `scalegrid_clusters` | List clusters, optionally filtered by database type. |
| `scalegrid_cloud_profile` | Look up a cloud profile by ID or name. |

All resources support `terraform import`.

## Usage

```terraform
terraform {
  required_providers {
    scalegrid = {
      source  = "requestflo/scalegrid"
      version = "~> 0.1"
    }
  }
}

provider "scalegrid" {
  email   = "you@example.com"
  api_key = var.scalegrid_api_key
}

resource "scalegrid_cloud_profile" "aws" {
  name           = "aws-production"
  cloud_provider = "aws"
  region         = "us-east-1"
  access_key     = var.aws_access_key
  secret_key     = var.aws_secret_key
}

resource "scalegrid_cluster" "mongo" {
  name             = "production-mongo"
  database_type    = "mongodb"
  version          = "7.0"
  deployment_type  = "replicaset"
  cloud_profile_id = scalegrid_cloud_profile.aws.id
  region           = "us-east-1"
  size_id          = "AWS_M5_LARGE"
  disk_size_gb     = 100
}

resource "scalegrid_firewall_rule" "office" {
  cluster_id = scalegrid_cluster.mongo.id
  cidr       = "203.0.113.0/24"
}
```

More examples live under [`examples/`](./examples) and reference docs under
[`docs/`](./docs).

## Authentication

Generate an API key in the ScaleGrid console (Account → API Keys). Configuration
can come from the provider block or environment variables:

| Argument    | Environment variable  | Default |
|-------------|-----------------------|---------|
| `email`     | `SCALEGRID_EMAIL`     | — |
| `api_key`   | `SCALEGRID_API_KEY`   | — |
| `auth_mode` | `SCALEGRID_AUTH_MODE` | `basic` |
| `base_url`  | `SCALEGRID_BASE_URL`  | `https://api.scalegrid.io/v1` |

`auth_mode` selects between `basic` (email + API key via HTTP Basic auth) and
`bearer` (API key as a bearer token).

## API endpoint assumptions

ScaleGrid's public API reference is gated behind login, so this provider models
the API surface against ScaleGrid's documented resource model. The two pieces
most likely to need adjustment for a specific account are centralized so they can
be changed without touching resource logic:

- **Base URL** — set `base_url` (or `SCALEGRID_BASE_URL`) to match the endpoint
  your account uses.
- **Authentication scheme** — switch `auth_mode` between `basic` and `bearer`.

The concrete request paths live in [`internal/client`](./internal/client)
(`/clusters`, `/clusters/{id}/firewall_rules`, `/cloud_profiles`,
`/clusters/{id}/backups`, `/jobs/{id}`). They are isolated in that package so
the Terraform-facing schema does not change if a path needs tweaking. If you
have access to the authoritative API reference, verify these against it.

## Development

Requires Go 1.23+.

```sh
make build      # compile the provider
make test       # run unit tests
make vet        # go vet
make fmt        # gofmt
make install    # build + install into ~/.terraform.d/plugins for local testing
```

To use a locally built provider, add a [dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)
to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/requestflo/scalegrid" = "/path/to/go/bin"
  }
  direct {}
}
```

### Testing

Unit tests cover the API client (request shaping, auth headers, error handling,
polling) and validate every resource/data-source schema:

```sh
go test ./...
```

Acceptance tests (`make testacc`) run against a live ScaleGrid account and create
real, billable resources. They are gated behind `TF_ACC=1`.

## Repository layout

```
.
├── main.go                     # provider entrypoint
├── internal/
│   ├── client/                 # ScaleGrid REST API client (no Terraform deps)
│   └── provider/               # provider, resources, and data sources
├── docs/                       # registry documentation
├── examples/                   # example Terraform configurations
└── .github/workflows/          # CI and release pipelines
```
