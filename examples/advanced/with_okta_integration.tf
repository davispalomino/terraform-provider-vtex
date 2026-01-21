# Advanced example: Integration with Okta data
# This example shows how to integrate the provider with existing data

terraform {
  required_providers {
    vtex = {
      source  = "registry.terraform.io/davispalomino/vtex"
      version = "0.1.0"
    }
    okta = {
      source  = "okta/okta"
      version = "~> 4.0"
    }
  }
}

provider "vtex" {
  vtex_base_url   = "https://vendor.myvtex.com"
  okta_url        = var.okta_url
  okta_client_id  = var.okta_client_id
  okta_secret     = var.okta_secret
  okta_grant_type = var.okta_grant_type
  okta_scope      = var.okta_scope
}

# Variables
variable "okta_url" {
  type = string
}

variable "okta_client_id" {
  type      = string
  sensitive = true
}

variable "okta_secret" {
  type      = string
  sensitive = true
}

variable "okta_grant_type" {
  type = string
}

variable "okta_scope" {
  type = string
}

# Okta groups to VTEX roles mapping
locals {
  vtex_role_mapping = {
    "vendor_admin"   = "Owner"
    "vendor_content" = "Content"
  }

  # Example of how to define users
  # In production, this would come from a data source like Okta
  vtex_users = {
    "user1.vtex_operations.prd" = {
      email  = "user1@email.com"
      env    = "prd"
      role   = "vendor_admin"
      status = "ACTIVE"
    }
    "user2.vtex_finance.prd" = {
      email  = "user2@email.com"
      env    = "prd"
      role   = "vendor_content"
      status = "ACTIVE"
    }
  }

  # Environment to VTEX account mapping
  env_to_account = {
    "prd"     = "vendorprd"
    "stg"     = "vendorstg"
    "qa"      = "vendorqa"
    "seller1" = "seller1"
    "seller2" = "seller2"
  }
}

# Create VTEX users based on config
resource "vtex_user_role" "users" {
  for_each = {
    for k, v in local.vtex_users : k => v
    if v.status == "ACTIVE"
  }

  email     = each.value.email
  account   = local.env_to_account[each.value.env]
  role_name = local.vtex_role_mapping[each.value.role]
}
