terraform {
  required_providers {
    scalegrid = {
      source  = "requestflo/scalegrid"
      version = "~> 0.1"
    }
  }
}

provider "scalegrid" {
  # Credentials can also be supplied via the SCALEGRID_EMAIL and
  # SCALEGRID_API_KEY environment variables.
  email   = "you@example.com"
  api_key = var.scalegrid_api_key
}
