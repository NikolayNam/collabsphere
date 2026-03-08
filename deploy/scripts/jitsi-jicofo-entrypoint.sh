#!/bin/sh
set -eu

read_file_secret() {
  file_path="$1"
  tr -d '\r\n' < "$file_path"
}

export ENABLE_AUTH=1
export AUTH_TYPE=jwt
export ENABLE_AUTO_OWNER=0
export XMPP_DOMAIN=meet.jitsi
export XMPP_AUTH_DOMAIN=auth.meet.jitsi
export XMPP_INTERNAL_MUC_DOMAIN=internal-muc.meet.jitsi
export XMPP_MUC_DOMAIN=muc.meet.jitsi
export XMPP_SERVER=xmpp.meet.jitsi
export XMPP_PORT=5222
export JICOFO_AUTH_USER=focus
export JICOFO_AUTH_PASSWORD="$(read_file_secret "${JITSI_JICOFO_AUTH_PASSWORD_FILE}")"
export JICOFO_COMPONENT_SECRET="$(read_file_secret "${JITSI_JICOFO_COMPONENT_SECRET_FILE}")"
export JVB_BREWERY_MUC=jvbbrewery
export JVB_XMPP_AUTH_DOMAIN=auth.meet.jitsi
export JVB_XMPP_INTERNAL_MUC_DOMAIN=internal-muc.meet.jitsi
export JVB_XMPP_SERVER=xmpp.meet.jitsi
export JVB_XMPP_PORT=5222

exec /init
