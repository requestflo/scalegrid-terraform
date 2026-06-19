# An AWS cloud profile for Bring Your Own Cloud deployments.
resource "scalegrid_cloud_profile" "aws" {
  name           = "aws-production"
  cloud_provider = "aws"
  region         = "us-east-1"
  access_key     = var.aws_access_key
  secret_key     = var.aws_secret_key
}

# An Azure cloud profile.
resource "scalegrid_cloud_profile" "azure" {
  name            = "azure-production"
  cloud_provider  = "azure"
  region          = "eastus"
  subscription_id = var.azure_subscription_id
  tenant_id       = var.azure_tenant_id
  client_id       = var.azure_client_id
  client_secret   = var.azure_client_secret
}
