---
page_title: "scalegrid_cloud_profile Data Source - terraform-provider-scalegrid"
description: |-
  Fetches a single ScaleGrid cloud profile by ID or name.
---

# scalegrid_cloud_profile (Data Source)

Fetches a single ScaleGrid cloud profile by ID or name.

## Example Usage

```terraform
data "scalegrid_cloud_profile" "aws" {
  name = "aws-production"
}
```

## Schema

### Optional

- `id` (String) ID of the cloud profile. Either `id` or `name` must be set.
- `name` (String) Name of the cloud profile. Either `id` or `name` must be set.

### Read-Only

- `cloud_provider` (String)
- `region` (String)
- `created_at` (String)
