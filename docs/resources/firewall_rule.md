---
page_title: "scalegrid_firewall_rule Resource - terraform-provider-scalegrid"
description: |-
  Manages a firewall rule that allows a CIDR block to connect to a ScaleGrid cluster.
---

# scalegrid_firewall_rule (Resource)

Manages a firewall rule that allows a CIDR block to connect to a ScaleGrid
cluster. Firewall rules apply to clusters that are open to the internet.

## Example Usage

```terraform
resource "scalegrid_firewall_rule" "office" {
  cluster_id  = scalegrid_cluster.mongo.id
  cidr        = "203.0.113.0/24"
  description = "Office network"
}
```

## Schema

### Required

- `cluster_id` (String) ID of the cluster the rule applies to. Forces replacement.
- `cidr` (String) CIDR block allowed to connect.

### Optional

- `description` (String) Optional human-readable description.

### Read-Only

- `id` (String) Unique identifier of the firewall rule.

## Import

```shell
terraform import scalegrid_firewall_rule.office <cluster_id>:<rule_id>
```
