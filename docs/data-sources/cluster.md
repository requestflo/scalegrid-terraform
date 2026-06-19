---
page_title: "scalegrid_cluster Data Source - terraform-provider-scalegrid"
description: |-
  Fetches a single ScaleGrid cluster by ID.
---

# scalegrid_cluster (Data Source)

Fetches a single ScaleGrid cluster by ID.

## Example Usage

```terraform
data "scalegrid_cluster" "existing" {
  id = "5f8d0c9b1c9d440000a1b2c3"
}
```

## Schema

### Required

- `id` (String) ID of the cluster to look up.

### Read-Only

- `name` (String)
- `database_type` (String)
- `version` (String)
- `deployment_type` (String)
- `cloud_profile_id` (String)
- `region` (String)
- `size_id` (String)
- `disk_size_gb` (Number)
- `status` (String)
- `host` (String)
- `port` (Number)
- `connection_string` (String, Sensitive)
- `created_at` (String)
