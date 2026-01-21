# Terraform Provider VTEX

Terraform provider to manage users and roles in VTEX.

## Important: Prerequisites

> **WARNING: VTEX Apps Service Required**
>
> Before using this provider, you must install a VTEX Apps Service in your VTEX account.
> This app provides the endpoints needed to add or remove users.
>
> **Endpoints required:**
> - `/_v/create-user-role` - To add users
> - `/_v/remove-user-role` - To remove users
>
> **Without this app installed, the provider will NOT work.**

## Requirements

- [Go](https://golang.org/doc/install) >= 1.21
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

## Build the Provider

```bash
# Download dependencies
make deps

# Build and install locally
make install
```

## Usage

```hcl
terraform {
  required_providers {
    vtex = {
      source  = "registry.terraform.io/davispalomino/vtex"
      version = "0.1.0"
    }
  }
}

provider "vtex" {
  vtex_base_url   = "https://vendor.myvtex.com"
  okta_url        = var.okta_url
  okta_client_id  = var.okta_client_id
  okta_secret     = var.okta_secret
  okta_grant_type = "authorization_code"
  okta_scope      = "scope_vendor"
}

resource "vtex_user_role" "example" {
  email     = "user@example.com"
  account   = "vendor"
  role_name = "Operation"
}
```

## Available Resources

### vtex_user_role

Manages a user with a specific role in a VTEX account.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `email` | string | Yes | User email |
| `name` | string | No | User name (if not given, it is taken from email) |
| `account` | string | Yes | VTEX account (e.g. vendor) |
| `role_name` | string | Yes | Role name (e.g. Owner, Operation) |

#### Exported Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | string | Unique ID (email:account:role_name) |

#### Import

```bash
terraform import vtex_user_role.example "email@example.com:account:role_name"
```

## Features

- **Token caching**: The provider reuses tokens until they expire
- **Auto token renewal**: If a token expires, a new one is requested
- **Retries with backoff**: Up to 20 retries with exponential backoff
- **Rate limit handling**: Waits and retries on 429, 404, 504 errors
- **Sensitive data protection**: Okta credentials are marked as sensitive

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Clean builds
make clean
```

## Project Structure

```
.
├── main.go                           # Entry point
├── go.mod                            # Go dependencies
├── Makefile                          # Build commands
├── internal/
│   ├── provider/
│   │   ├── provider.go               # Provider config
│   │   └── vtex_user_role_resource.go # vtex_user_role resource
│   └── client/
│       └── client.go                 # HTTP client for VTEX API
└── examples/
    ├── basic/main.tf                 # Basic example
    └── advanced/with_okta_integration.tf # Advanced example
```
