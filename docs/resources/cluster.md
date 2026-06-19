---
page_title: "scalegrid_cluster Resource - terraform-provider-scalegrid"
description: |-
  Manages a ScaleGrid database deployment (cluster).
---

# scalegrid_cluster (Resource)

Manages a ScaleGrid database deployment (cluster) for MongoDB, Redis, MySQL, or
PostgreSQL. Provisioning is asynchronous; Terraform waits for the cluster to
become available before completing the apply.

## Example Usage

```terraform
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
```

## Schema

### Required

- `name` (String) Human-readable name of the cluster.
- `database_type` (String) Database engine: `mongodb`, `redis`, `mysql`, or `postgresql`. Forces replacement.
- `cloud_profile_id` (String) ID of the cloud profile used to provision the cluster. Forces replacement.
- `size_id` (String) Instance size identifier. Changing this scales the cluster in place.

### Optional

- `version` (String) Engine version to deploy. Forces replacement.
- `deployment_type` (String) Topology: `standalone`, `replicaset`, `sharded`, or `cluster`. Forces replacement.
- `region` (String) Cloud region. Forces replacement.
- `disk_size_gb` (Number) Disk size in GB.
- `shard_count` (Number) Number of shards (sharded deployments only). Forces replacement.
- `ssl_enabled` (Boolean) Whether SSL/TLS is enabled. Defaults to `true`. Forces replacement.
- `encryption_at_rest` (Boolean) Whether encryption at rest is enabled. Defaults to `false`. Forces replacement.
- `tags` (List of String) Optional tags applied to the cluster.

### Read-Only

- `id` (String) Unique identifier of the cluster.
- `status` (String) Current lifecycle status.
- `host` (String) Primary hostname for connecting to the cluster.
- `port` (Number) Connection port.
- `connection_string` (String, Sensitive) Connection string.
- `created_at` (String) Creation timestamp.

## Import

```shell
terraform import scalegrid_cluster.mongo <cluster_id>
```
