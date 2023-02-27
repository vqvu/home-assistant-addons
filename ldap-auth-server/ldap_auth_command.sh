#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

HOST="${1:-localhost:8080}"
ESCAPED_USERNAME="$(jq --raw-input <<< "${username}")"
ESCAPED_PASSWORD="$(jq --raw-input <<< "${password}")"
curl http://${HOST}/hass_authenticate \
    -X POST \
    -H "Content-Type: application/json" \
    --fail-with-body \
    -d "$(cat << EOF
{"username": ${ESCAPED_USERNAME},"password": ${ESCAPED_PASSWORD}}
EOF
)"
