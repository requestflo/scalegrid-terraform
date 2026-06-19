# Allow an office network to reach the cluster.
resource "scalegrid_firewall_rule" "office" {
  cluster_id  = scalegrid_cluster.mongo.id
  cidr        = "203.0.113.0/24"
  description = "Office network"
}
