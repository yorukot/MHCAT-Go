#!/usr/bin/env sh
set -eu

: "${MHCAT_DISCORD_TOKEN:?MHCAT_DISCORD_TOKEN is required}"
: "${MHCAT_MONGODB_URI:?MHCAT_MONGODB_URI is required}"
: "${MHCAT_MONGODB_DATABASE:?MHCAT_MONGODB_DATABASE is required}"

if [ "${MHCAT_STAGING_MODE:-false}" != "true" ]; then
  echo "refusing gateway smoke: MHCAT_STAGING_MODE=true is required" >&2
  exit 1
fi
if [ "${MHCAT_STAGING_ALLOW_GATEWAY_SMOKE:-false}" != "true" ]; then
  echo "refusing gateway smoke: MHCAT_STAGING_ALLOW_GATEWAY_SMOKE=true is required" >&2
  exit 1
fi

MHCAT_DISCORD_ENABLE_GATEWAY=true \
MHCAT_DISCORD_GATEWAY_SMOKE_TEST=true \
go run ./cmd/mhcat-bot
