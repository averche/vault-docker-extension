#!/bin/sh

set -e

export VAULT_ADDR='http://127.0.0.1:8201'

vault server -dev -dev-listen-address="0.0.0.0:8201" &
sleep 5s

vault login -no-print root

vault kv put -mount=secret sample-secret1 "password=trustno1"
vault kv put -mount=secret sample-secret2 "password=bad-password"
vault kv put -mount=secret sample-secret3 "password=abc123"

# This container is now healthy
touch /tmp/healthy

# Keep container alive
tail -f /dev/null & trap 'kill %1' TERM ; wait
