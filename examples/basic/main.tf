terraform {
  required_providers {
    vtex = {
      source  = "registry.terraform.io/davispalomino/vtex"
      version = "0.1.0"
    }
  }
}

# Provider config
provider "vtex" {
  vtex_base_url   = "https://vendor.myvtex.com"
  okta_url        = var.okta_url
  okta_client_id  = var.okta_client_id
  okta_secret     = var.okta_secret
  okta_grant_type = var.okta_grant_type
  okta_scope      = var.okta_scope
}

# Sensitive variables
variable "okta_url" {
  description = "Okta OAuth2 endpoint URL"
  type        = string
}

variable "okta_client_id" {
  description = "Okta Client ID"
  type        = string
  sensitive   = true
}

variable "okta_secret" {
  description = "Okta Client Secret"
  type        = string
  sensitive   = true
}

variable "okta_grant_type" {
  description = "OAuth2 grant type"
  type        = string
}

variable "okta_scope" {
  description = "OAuth2 scope"
  type        = string
}

# Example 1: Single user
resource "vtex_user_role" "admin_user" {
  email     = "admin@email.com"
  account   = "seller1"
  role_name = "Owner"
}

# Example 2: User with custom name
resource "vtex_user_role" "operations_user" {
  email     = "operator@email.com"
  name      = "Davis Operator"
  account   = "seller1"
  role_name = "Operation"
}

# Example 3: Multiple users using for_each
variable "vtex_users" {
  description = "Map of VTEX users"
  type = map(object({
    email     = string
    account   = string
    role_name = string
  }))
  default = {
    "user1-terraform" = {
      email     = "user1@email.com"
      account   = "vendor"
      role_name = "Owner"
    }
    "user2-terraform" = {
      email     = "user2@email.com"
      account   = "vendor"
      role_name = "Operation"
    }
  }
}

resource "vtex_user_role" "bulk_users" {
  for_each = var.vtex_users

  email     = each.value.email
  account   = each.value.account
  role_name = each.value.role_name
}

# Outputs
output "admin_user_id" {
  value = vtex_user_role.admin_user.id
}

output "all_users" {
  value = { for k, v in vtex_user_role.bulk_users : k => v.id }
}
