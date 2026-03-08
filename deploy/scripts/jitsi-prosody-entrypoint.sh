#!/bin/sh
set -eu

read_file_secret() {
  file_path="$1"
  tr -d '\r\n' < "$file_path"
}

export JWT_APP_SECRET="$(read_file_secret "${JITSI_APP_SECRET_FILE}")"
export JICOFO_AUTH_PASSWORD="$(read_file_secret "${JITSI_JICOFO_AUTH_PASSWORD_FILE}")"
export JICOFO_COMPONENT_SECRET="$(read_file_secret "${JITSI_JICOFO_COMPONENT_SECRET_FILE}")"
export JVB_AUTH_USER=jvb
export JVB_AUTH_PASSWORD="$(read_file_secret "${JITSI_JVB_AUTH_PASSWORD_FILE}")"
export ENABLE_AUTH=1
export ENABLE_GUESTS=0
export AUTH_TYPE=jwt
export JWT_APP_ID="${JITSI_APP_ID}"
export JWT_ACCEPTED_ISSUERS="${JITSI_ISSUER}"
export JWT_ACCEPTED_AUDIENCES="${JITSI_AUDIENCE}"
export XMPP_DOMAIN=meet.jitsi
export XMPP_AUTH_DOMAIN=auth.meet.jitsi
export XMPP_GUEST_DOMAIN=guest.meet.jitsi
export XMPP_MUC_DOMAIN=muc.meet.jitsi
export XMPP_INTERNAL_MUC_DOMAIN=internal-muc.meet.jitsi
export XMPP_HIDDEN_DOMAIN=recorder.meet.jitsi
export XMPP_RECORDER_DOMAIN=recorder.meet.jitsi
export XMPP_PORT=5222

exec /init
