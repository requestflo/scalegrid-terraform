---
page_title: "ScaleGrid Provider"
description: |-
  The ScaleGrid provider manages database deployments (MongoDB, Redis, MySQL, PostgreSQL) through the ScaleGrid REST API.
---

# ScaleGrid Provider

The ScaleGrid provider is used to manage [ScaleGrid](https://scalegrid.io) database
deployments and related resources (cloud profiles, firewall rules, backups)
through the [ScaleGrid REST API](https://scalegrid.io/api/).

## Example Usage

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
```

## Authentication

Generate an API key from the ScaleGrid console (Account → API Keys). The provider
supports two authentication schemes, controlled by `auth_mode`:

* `basic` (default) — HTTP Basic authentication using your account `email` as the
  username and the `api_key` as the password.
* `bearer` — sends the `api_key` as a bearer token.

Credentials may be supplied in the provider block or via environment variables:

| Argument    | Environment variable   |
|-------------|------------------------|
| `email`     | `SCALEGRID_EMAIL`      |
| `api_key`   | `SCALEGRID_API_KEY`    |
| `auth_mode` | `SCALEGRID_AUTH_MODE`  |
| `base_url`  | `SCALEGRID_BASE_URL`   |

## Schema

### Optional

- `base_url` (String) Base URL of the ScaleGrid API. Defaults to `https://api.scalegrid.io/v1`.
- `email` (String) ScaleGrid account email, used as the username for basic authentication.
- `api_key` (String, Sensitive) ScaleGrid API key generated from the console.
- `auth_mode` (String) Authentication scheme: `basic` (default) or `bearer`.
