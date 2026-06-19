---
page_title: "scalegrid_cloud_profile Resource - terraform-provider-scalegrid"
description: |-
  Manages a ScaleGrid cloud profile.
---

# scalegrid_cloud_profile (Resource)

Manages a ScaleGrid cloud profile, which stores the cloud credentials used to
provision clusters in a Bring Your Own Cloud account.

## Example Usage

```terraform
resource "scalegrid_cloud_profile" "aws" {
  name           = "aws-production"
  cloud_provider = "aws"
  region         = "us-east-1"
  access_key     = var.aws_access_key
  secret_key     = var.aws_secret_key
}
```

## Schema

### Required

- `name` (String) Human-readable name of the cloud profile.
- `cloud_provider` (String) One of `aws`, `azure`, `gcp`, `digitalocean`, `oci`, or `vmware`. Forces replacement.

### Optional

- `region` (String) Default cloud region for the profile.
- `access_key` (String) Access key (AWS) or equivalent identifier.
- `secret_key` (String, Sensitive) Secret key (AWS). Write-only.
- `subscription_id` (String) Subscription ID (Azure).
- `tenant_id` (String) Tenant ID (Azure).
- `client_id` (String) Client/application ID (Azure).
- `client_secret` (String, Sensitive) Client secret (Azure). Write-only.

### Read-Only

- `id` (String) Unique identifier of the cloud profile.
- `created_at` (String) Creation timestamp.

## Import

```shell
terraform import scalegrid_cloud_profile.aws <cloud_profile_id>
```
