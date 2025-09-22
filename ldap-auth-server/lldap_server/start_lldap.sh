#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

# Determine the script directory regardless of CWD
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TLS_DIR="${SCRIPT_DIR}/data/tls"
mkdir -p "${TLS_DIR}"

# File paths
CA_KEY="${TLS_DIR}/rootCA.key.pem"
CA_CERT="${TLS_DIR}/rootCA.cert.pem"
SERVER_KEY="${TLS_DIR}/server.key.pem"
SERVER_CSR="${TLS_DIR}/server.csr.pem"
SERVER_CERT="${TLS_DIR}/server.cert.pem"
SERVER_CHAIN="${TLS_DIR}/server.chain.pem"
EXTFILE="${TLS_DIR}/server_cert_ext.cnf"

# Create default extfile for server cert SANs if not present
if [[ ! -f "${EXTFILE}" ]]; then
  cat > "${EXTFILE}" <<'EOF'
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment, keyAgreement
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = lldap
IP.1 = 127.0.0.1
IP.2 = 172.17.0.2
EOF
fi

# Generate Root CA if missing
if [[ ! -f "${CA_KEY}" || ! -f "${CA_CERT}" ]]; then
  echo "Generating Root CA..."
  # Generate private key for CA
  openssl genrsa -out "${CA_KEY}" 4096

  # Self-signed CA certificate (10 years)
  openssl req -x509 -new -nodes \
    -key "${CA_KEY}" \
    -sha256 -days 3650 \
    -subj "/C=US/ST=Local/L=Local/O=Dev/OU=Dev/CN=Local Dev Root CA" \
    -out "${CA_CERT}"
else
  echo "Root CA already exists at ${CA_CERT}"
fi

# Generate server key and CSR if missing
if [[ ! -f "${SERVER_KEY}" ]]; then
  echo "Generating server key..."
  openssl genrsa -out "${SERVER_KEY}" 4096
fi

if [[ ! -f "${SERVER_CSR}" ]]; then
  echo "Generating server CSR..."
  openssl req -new \
    -key "${SERVER_KEY}" \
    -subj "/C=US/ST=Local/L=Local/O=Dev/OU=Dev/CN=localhost" \
    -out "${SERVER_CSR}"
fi

# Sign server cert with Root CA if missing or expired
NEED_SIGN="true"
if [[ -f "${SERVER_CERT}" ]]; then
  # Basic check: ensure cert is not expired (1 day buffer)
  if openssl x509 -checkend 86400 -noout -in "${SERVER_CERT}" >/dev/null 2>&1; then
    NEED_SIGN="false"
  fi
fi

if [[ "${NEED_SIGN}" == "true" ]]; then
  echo "Signing server certificate with Root CA..."
  openssl x509 -req -in "${SERVER_CSR}" \
    -CA "${CA_CERT}" -CAkey "${CA_KEY}" \
    -CAcreateserial \
    -sha256 -days 825 \
    -extfile "${EXTFILE}" \
    -out "${SERVER_CERT}"
fi

# Build certificate chain (server cert + CA cert)
cat "${SERVER_CERT}" "${CA_CERT}" > "${SERVER_CHAIN}"

echo "Certificates generated in ${TLS_DIR}:"
ls -1 "${TLS_DIR}"

# Start the server using docker compose from the script directory
(cd "${SCRIPT_DIR}" && docker compose up)

echo "Server started via docker compose."
