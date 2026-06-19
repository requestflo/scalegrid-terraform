# Trigger an on-demand backup of a cluster.
resource "scalegrid_backup" "snapshot" {
  cluster_id = scalegrid_cluster.mongo.id
}
