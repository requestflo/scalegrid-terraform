---
page_title: "scalegrid_clusters Data Source - terraform-provider-scalegrid"
description: |-
  Lists ScaleGrid clusters on the account, optionally filtered by database type.
---

# scalegrid_clusters (Data Source)

Lists ScaleGrid clusters on the account, optionally filtered by database type.

## Example Usage

```terraform
data "scalegrid_clusters" "postgres" {
  database_type = "postgresql"
}
```

## Schema

### Optional

- `database_type` (String) If set, only clusters with this database engine are returned.

### Read-Only

- `clusters` (List of Object) The list of clusters. Each element exposes the same
  read-only attributes as the `scalegrid_cluster` data source.
