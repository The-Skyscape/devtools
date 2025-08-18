# Application policy for Skyscape services

# Read-only access to integrations secrets
path "secret/data/integrations/*" {
  capabilities = ["read", "list"]
}

# Read-write access to workspace secrets
path "secret/data/workspaces/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Read-only access to application configuration
path "secret/data/config/*" {
  capabilities = ["read", "list"]
}

# Allow token renewal
path "auth/token/renew-self" {
  capabilities = ["update"]
}

# Allow token lookup
path "auth/token/lookup-self" {
  capabilities = ["read"]
}