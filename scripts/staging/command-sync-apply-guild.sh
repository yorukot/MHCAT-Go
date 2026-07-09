#!/usr/bin/env sh
set -eu

: "${MHCAT_DISCORD_TOKEN:?MHCAT_DISCORD_TOKEN is required}"
: "${MHCAT_DISCORD_APPLICATION_ID:?MHCAT_DISCORD_APPLICATION_ID is required}"
: "${MHCAT_STAGING_GUILD_ID:?MHCAT_STAGING_GUILD_ID is required}"

if [ "${MHCAT_STAGING_MODE:-false}" != "true" ]; then
  echo "refusing apply: MHCAT_STAGING_MODE=true is required" >&2
  exit 1
fi
if [ "${MHCAT_STAGING_ALLOW_COMMAND_APPLY:-false}" != "true" ]; then
  echo "refusing apply: MHCAT_STAGING_ALLOW_COMMAND_APPLY=true is required" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_SCOPE:-guild}" != "guild" ]; then
  echo "refusing apply: staging apply requires guild scope" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_ALLOW_DELETE:-false}" = "true" ]; then
  echo "refusing apply: command deletion is disabled for staging smoke" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE:-false}" = "true" ]; then
  echo "refusing apply: bulk overwrite is disabled for staging smoke" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TICKETS:-false}" = "true" ] && [ "${MHCAT_FEATURE_TICKETS_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: ticket command apply requires MHCAT_FEATURE_TICKETS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_POLLS:-false}" = "true" ] && [ "${MHCAT_FEATURE_POLLS_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: poll command apply requires MHCAT_FEATURE_POLLS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_QUERY_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: economy query command apply requires MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: economy sign-in commands apply requires MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: economy settings command apply requires MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WORK:-false}" = "true" ] && [ "${MHCAT_FEATURE_WORK_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: work command apply requires MHCAT_FEATURE_WORK_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS:-false}" = "true" ] && [ "${MHCAT_FEATURE_WARNINGS_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: warning-history command apply requires MHCAT_FEATURE_WARNINGS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS:-false}" = "true" ] && [ "${MHCAT_FEATURE_WARNING_SETTINGS_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: warning-settings command apply requires MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE:-false}" = "true" ] && [ "${MHCAT_FEATURE_TRANSLATE_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: translate command apply requires MHCAT_FEATURE_TRANSLATE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY:-false}" = "true" ] && [ "${MHCAT_FEATURE_BALANCE_QUERY_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: balance query command apply requires MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_REDEEM:-false}" = "true" ] && [ "${MHCAT_FEATURE_REDEEM_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: redeem command apply requires MHCAT_FEATURE_REDEEM_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: autochat config command apply requires MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: auto-notification config command apply requires MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: anti-scam config command apply requires MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: anti-scam report command apply requires MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT:-false}" = "true" ] && [ -z "${MHCAT_REPORT_WEBHOOK_URL:-}" ] && [ -z "${REPORT_WEBHOOK:-}" ]; then
  echo "refusing apply: anti-scam report command apply requires MHCAT_REPORT_WEBHOOK_URL or REPORT_WEBHOOK for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_LOGGING_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: logging config command apply requires MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST:-false}" = "true" ] && [ "${MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: gacha prize-list command apply requires MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND:-false}" = "true" ] && [ "${MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: lottery disabled command apply requires MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY:-false}" = "true" ] && [ "${MHCAT_FEATURE_STATS_QUERY_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: stats query command apply requires MHCAT_FEATURE_STATS_QUERY_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: birthday config command apply requires MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: announcement config command apply requires MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: announcement send command apply requires MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: text XP config command apply requires MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: voice XP config command apply requires MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS:-false}" = "true" ] && [ "${MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: XP profile disabled commands apply requires MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: voice-room config commands apply requires MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: join-role config command apply requires MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: welcome-message config command apply requires MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: verification config command apply requires MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW:-false}" = "true" ] && [ "${MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: verification flow command apply requires MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing apply: account-age config command apply requires MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi

echo "staging command sync apply: guild create/update only; no delete; no bulk overwrite" >&2
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TICKETS:-false}" = "true" ]; then
  echo "staging command sync apply: including ticket commands" >&2
else
  echo "staging command sync apply: ticket commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_POLLS:-false}" = "true" ]; then
  echo "staging command sync apply: including poll commands" >&2
else
  echo "staging command sync apply: poll commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY:-false}" = "true" ]; then
  echo "staging command sync apply: including economy query command" >&2
else
  echo "staging command sync apply: economy query command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN:-false}" = "true" ]; then
  echo "staging command sync apply: including economy sign-in commands" >&2
else
  echo "staging command sync apply: economy sign-in commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS:-false}" = "true" ]; then
  echo "staging command sync apply: including economy settings command" >&2
else
  echo "staging command sync apply: economy settings command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WORK:-false}" = "true" ]; then
  echo "staging command sync apply: including work command" >&2
else
  echo "staging command sync apply: work command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS:-false}" = "true" ]; then
  echo "staging command sync apply: including warning-history command" >&2
else
  echo "staging command sync apply: warning-history command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS:-false}" = "true" ]; then
  echo "staging command sync apply: including warning-settings command" >&2
else
  echo "staging command sync apply: warning-settings command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE:-false}" = "true" ]; then
  echo "staging command sync apply: including translate command" >&2
else
  echo "staging command sync apply: translate command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY:-false}" = "true" ]; then
  echo "staging command sync apply: including balance query command" >&2
else
  echo "staging command sync apply: balance query command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_REDEEM:-false}" = "true" ]; then
  echo "staging command sync apply: including redeem command" >&2
else
  echo "staging command sync apply: redeem command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including autochat config commands" >&2
else
  echo "staging command sync apply: autochat config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including auto-notification config commands" >&2
else
  echo "staging command sync apply: auto-notification config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including anti-scam config command" >&2
else
  echo "staging command sync apply: anti-scam config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT:-false}" = "true" ]; then
  echo "staging command sync apply: including anti-scam report command" >&2
else
  echo "staging command sync apply: anti-scam report command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including logging config command" >&2
else
  echo "staging command sync apply: logging config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST:-false}" = "true" ]; then
  echo "staging command sync apply: including gacha prize-list command" >&2
else
  echo "staging command sync apply: gacha prize-list command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND:-false}" = "true" ]; then
  echo "staging command sync apply: including lottery disabled command" >&2
else
  echo "staging command sync apply: lottery disabled command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY:-false}" = "true" ]; then
  echo "staging command sync apply: including stats query command" >&2
else
  echo "staging command sync apply: stats query command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including birthday config command" >&2
else
  echo "staging command sync apply: birthday config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including announcement config command" >&2
else
  echo "staging command sync apply: announcement config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND:-false}" = "true" ]; then
  echo "staging command sync apply: including announcement send command" >&2
else
  echo "staging command sync apply: announcement send command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including text XP config commands" >&2
else
  echo "staging command sync apply: text XP config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including voice XP config commands" >&2
else
  echo "staging command sync apply: voice XP config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS:-false}" = "true" ]; then
  echo "staging command sync apply: including XP profile disabled commands" >&2
else
  echo "staging command sync apply: XP profile disabled commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including voice-room config commands" >&2
else
  echo "staging command sync apply: voice-room config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including join-role config commands" >&2
else
  echo "staging command sync apply: join-role config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including welcome-message config commands" >&2
else
  echo "staging command sync apply: welcome-message config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including verification config command" >&2
else
  echo "staging command sync apply: verification config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW:-false}" = "true" ]; then
  echo "staging command sync apply: including verification flow command" >&2
else
  echo "staging command sync apply: verification flow command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG:-false}" = "true" ]; then
  echo "staging command sync apply: including account-age config command" >&2
else
  echo "staging command sync apply: account-age config command is excluded" >&2
fi

MHCAT_COMMAND_SYNC_SCOPE=guild \
MHCAT_COMMAND_SYNC_GUILD_ID="$MHCAT_STAGING_GUILD_ID" \
MHCAT_COMMAND_SYNC_ALLOW_DELETE=false \
MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE=false \
go run ./cmd/mhcat-command-sync \
  --scope guild \
  --guild-id "$MHCAT_STAGING_GUILD_ID" \
  --apply
