---
page_title: "scalegrid_backup Resource - terraform-provider-scalegrid"
description: |-
  Triggers and manages an on-demand backup of a ScaleGrid cluster.
---

# scalegrid_backup (Resource)

Triggers and manages an on-demand backup of a ScaleGrid cluster.

## Example Usage

```terraform
resource "scalegrid_backup" "snapshot" {
  cluster_id = scalegrid_cluster.mongo.id
}
```

## Schema

### Required

- `cluster_id` (String) ID of the cluster to back up. Forces replacement.

### Read-Only

- `id` (String) Unique identifier of the backup.
- `status` (String) Status of the backup.
- `size_bytes` (Number) Size of the backup in bytes.
- `type` (String) Type of the backup.
- `created_at` (String) Creation timestamp.

## Import

```shell
terraform import scalegrid_backup.snapshot <cluster_id>:<backup_id>
```
