# A MongoDB replica set deployed through a cloud profile.
resource "scalegrid_cluster" "mongo" {
  name             = "production-mongo"
  database_type    = "mongodb"
  version          = "7.0"
  deployment_type  = "replicaset"
  cloud_profile_id = scalegrid_cloud_profile.aws.id
  region           = "us-east-1"
  size_id          = "AWS_M5_LARGE"
  disk_size_gb     = 100
  ssl_enabled      = true

  tags = ["team:platform", "env:prod"]
}

output "mongo_connection_string" {
  value     = scalegrid_cluster.mongo.connection_string
  sensitive = true
}
