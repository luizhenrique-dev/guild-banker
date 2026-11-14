#!/usr/bin/env bash
# Runs inside the Keycloak container, after its readiness check.
set -euo pipefail

if [[ "${1:-}" != "--apply" ]]; then
    timeout --signal=TERM --kill-after=5s 60s bash "$0" --apply
    exit $?
fi

export PATH="/opt/keycloak/bin:$PATH"
umask 077
session_dir=$(mktemp -d)
trap 'rm -rf -- "$session_dir"' EXIT
step='authenticate administrator'
trap 'printf "Keycloak theme: unable to %s\n" "$step" >&2' ERR

# The CLI reads this variable without putting the password in process arguments.
export KC_CLI_PASSWORD="${KC_BOOTSTRAP_ADMIN_PASSWORD:?Missing admin password}"
kcadm.sh config credentials --config "$session_dir/session" \
    --server http://localhost:8080 --realm master \
    --user "${KC_BOOTSTRAP_ADMIN_USERNAME:?Missing admin username}" >/dev/null 2>&1
unset KC_CLI_PASSWORD

step='read guild-banker login theme'
current=$(kcadm.sh get realms/guild-banker --config "$session_dir/session" \
    --fields loginTheme --format csv --noquotes 2>/dev/null)
if [[ "$current" == 'guild-banker' ]]; then
    printf 'Keycloak theme: guild-banker already selected\n'
    exit 0
fi

step='select guild-banker login theme'
# --no-merge sends only the field owned by this setup, not a realm snapshot.
kcadm.sh update realms/guild-banker --config "$session_dir/session" \
    --no-merge -s loginTheme=guild-banker >/dev/null 2>&1

step='confirm guild-banker login theme'
selected=$(kcadm.sh get realms/guild-banker --config "$session_dir/session" \
    --fields loginTheme --format csv --noquotes 2>/dev/null)
if [[ "$selected" != 'guild-banker' ]]; then
    printf 'Keycloak theme: selection could not be confirmed\n' >&2
    exit 1
fi
printf 'Keycloak theme: guild-banker selected and confirmed\n'
