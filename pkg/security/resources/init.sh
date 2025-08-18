#!/bin/bash
# Initialize Vault for first use

set -e

VAULT_ADDR="${VAULT_ADDR:-http://localhost:8200}"
VAULT_CONTAINER="${VAULT_CONTAINER:-vault}"

echo "Initializing Vault..."

# Check if already initialized
INIT_STATUS=$(docker exec $VAULT_CONTAINER vault status -format=json 2>/dev/null | jq -r '.initialized' || echo "false")

if [ "$INIT_STATUS" = "false" ]; then
    echo "Vault not initialized, initializing now..."
    
    # Initialize Vault
    docker exec $VAULT_CONTAINER vault operator init \
        -key-shares=5 \
        -key-threshold=3 \
        -format=json > /tmp/vault-init.json
    
    echo "Vault initialized successfully!"
    echo "Keys saved to /tmp/vault-init.json"
    
    # Auto-unseal with first 3 keys
    echo "Unsealing Vault..."
    UNSEAL_KEY_1=$(cat /tmp/vault-init.json | jq -r '.unseal_keys_b64[0]')
    UNSEAL_KEY_2=$(cat /tmp/vault-init.json | jq -r '.unseal_keys_b64[1]')
    UNSEAL_KEY_3=$(cat /tmp/vault-init.json | jq -r '.unseal_keys_b64[2]')
    
    docker exec $VAULT_CONTAINER vault operator unseal $UNSEAL_KEY_1
    docker exec $VAULT_CONTAINER vault operator unseal $UNSEAL_KEY_2
    docker exec $VAULT_CONTAINER vault operator unseal $UNSEAL_KEY_3
    
    # Get root token
    ROOT_TOKEN=$(cat /tmp/vault-init.json | jq -r '.root_token')
    
    # Enable KV secrets engine
    echo "Configuring secrets engine..."
    docker exec -e VAULT_TOKEN=$ROOT_TOKEN $VAULT_CONTAINER \
        vault secrets enable -path=secret kv-v2
    
    echo "Vault setup complete!"
    echo "Root token: $ROOT_TOKEN"
else
    echo "Vault is already initialized"
fi