#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

readonly HOST="${1:-localhost:8080}"
ESCAPED_USERNAME="$(jq --raw-input <<< "${username}")"
ESCAPED_PASSWORD="$(jq --raw-input <<< "${password}")"
readonly ESCAPED_USERNAME
readonly ESCAPED_PASSWORD

if curl "http://${HOST}/hass_authenticate" \
    -X POST \
    -H "Content-Type: application/json" \
    --fail-with-body \
    --silent \
    --show-error \
    -d "$(cat << EOF
{"username": ${ESCAPED_USERNAME},"password": ${ESCAPED_PASSWORD}}
EOF
)"; then
  # Hard-coded metadata
  #
  # This block runs if the authentication succeeds.
  # You may hard-code any values supported by the command-line provider here.
  # See https://www.home-assistant.io/docs/authentication/providers/#command-line
  #
  # The default marks accounts as regular users and local only to avoid
  # unintentionally granting new accounts extra privileges.
  echo "group = system-users"
  echo "local_only = true"

  # Messages printed to stderr show up in the Home Assistant logs. You can use
  # the presence of this message to write automation if you'd like. Or remove
  # it if you'd prefer to avoid the logging.
  echo "LDAP Auth Success - User: ${username}" >&2
  exit 0
else
  echo "LDAP Auth Failed - User: ${username}" >&2
  exit 1
fi
