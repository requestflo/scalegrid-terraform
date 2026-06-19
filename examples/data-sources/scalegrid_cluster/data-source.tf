# Look up a single cluster by ID.
data "scalegrid_cluster" "existing" {
  id = "5f8d0c9b1c9d440000a1b2c3"
}

# List all PostgreSQL clusters on the account.
data "scalegrid_clusters" "postgres" {
  database_type = "postgresql"
}

# Resolve a cloud profile by name.
data "scalegrid_cloud_profile" "aws" {
  name = "aws-production"
}
