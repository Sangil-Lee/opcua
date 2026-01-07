#!/bin/bash
# Setup PKI (Public Key Infrastructure) for OPC UA server

set -e

echo "Setting up PKI for OPC UA server..."

# Create PKI directory structure
mkdir -p pki

# Generate private key
openssl genrsa -out pki/server.key 2048

# Generate self-signed certificate
openssl req -new -x509 -key pki/server.key -out pki/server.crt -days 365 \
  -subj "/C=US/ST=State/L=City/O=go-opcua-sim/CN=localhost"

echo "✓ PKI setup complete"
echo ""
echo "Created files:"
echo "  - pki/server.key (private key)"
echo "  - pki/server.crt (certificate)"
echo ""
