#!/bin/sh
set -eu

read_file_secret() {
  file_path="$1"
  tr -d '\r\n' < "$file_path"
}

export XMPP_AUTH_DOMAIN=auth.meet.jitsi
export XMPP_INTERNAL_MUC_DOMAIN=internal-muc.meet.jitsi
export XMPP_SERVER=xmpp.meet.jitsi
export XMPP_PORT=5222
export JVB_AUTH_USER=jvb
export JVB_AUTH_PASSWORD="$(read_file_secret "${JITSI_JVB_AUTH_PASSWORD_FILE}")"
export JVB_BREWERY_MUC=jvbbrewery
export JVB_PORT="${JITSI_JVB_PORT}"
export JVB_ADVERTISE_IPS="${JITSI_JVB_ADVERTISE_IPS}"
export PUBLIC_URL="${JITSI_PUBLIC_URL}"

exec /init
