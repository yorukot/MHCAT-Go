#!/usr/bin/env sh
set -eu

: "${MHCAT_DISCORD_TOKEN:?MHCAT_DISCORD_TOKEN is required}"
: "${MHCAT_DISCORD_APPLICATION_ID:?MHCAT_DISCORD_APPLICATION_ID is required}"
: "${MHCAT_STAGING_GUILD_ID:?MHCAT_STAGING_GUILD_ID is required}"

if [ "${MHCAT_COMMAND_SYNC_SCOPE:-guild}" != "guild" ]; then
  echo "refusing command sync: staging dry-run requires guild scope" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TICKETS:-false}" = "true" ] && [ "${MHCAT_FEATURE_TICKETS_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: ticket command dry-run requires MHCAT_FEATURE_TICKETS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_POLLS:-false}" = "true" ] && [ "${MHCAT_FEATURE_POLLS_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: poll command dry-run requires MHCAT_FEATURE_POLLS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_QUERY_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: economy query command dry-run requires MHCAT_FEATURE_ECONOMY_QUERY_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: economy sign-in commands dry-run requires MHCAT_FEATURE_ECONOMY_SIGNIN_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: economy settings command dry-run requires MHCAT_FEATURE_ECONOMY_SETTINGS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: economy coin-admin command dry-run requires MHCAT_FEATURE_ECONOMY_COIN_ADMIN_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: economy coin-rank command dry-run requires MHCAT_FEATURE_ECONOMY_COIN_RANK_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_RPS_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: economy RPS command dry-run requires MHCAT_FEATURE_ECONOMY_RPS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE:-false}" = "true" ] && [ "${MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: economy profile command dry-run requires MHCAT_FEATURE_ECONOMY_PROFILE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WORK:-false}" = "true" ] && [ "${MHCAT_FEATURE_WORK_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: work command dry-run requires MHCAT_FEATURE_WORK_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS:-false}" = "true" ] && [ "${MHCAT_FEATURE_WARNINGS_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: warning-history command dry-run requires MHCAT_FEATURE_WARNINGS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS:-false}" = "true" ] && [ "${MHCAT_FEATURE_WARNING_SETTINGS_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: warning-settings command dry-run requires MHCAT_FEATURE_WARNING_SETTINGS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL:-false}" = "true" ] && [ "${MHCAT_FEATURE_WARNING_REMOVAL_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: warning-removal commands dry-run requires MHCAT_FEATURE_WARNING_REMOVAL_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE:-false}" = "true" ] && [ "${MHCAT_FEATURE_WARNING_ISSUE_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: warning issue command dry-run requires MHCAT_FEATURE_WARNING_ISSUE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP:-false}" = "true" ] && [ "${MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: message cleanup command dry-run requires MHCAT_FEATURE_MESSAGE_CLEANUP_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA:-false}" = "true" ] && [ "${MHCAT_FEATURE_DELETE_DATA_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: delete data command dry-run requires MHCAT_FEATURE_DELETE_DATA_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE:-false}" = "true" ] && [ "${MHCAT_FEATURE_TRANSLATE_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: translate command dry-run requires MHCAT_FEATURE_TRANSLATE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY:-false}" = "true" ] && [ "${MHCAT_FEATURE_BALANCE_QUERY_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: balance query command dry-run requires MHCAT_FEATURE_BALANCE_QUERY_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_REDEEM:-false}" = "true" ] && [ "${MHCAT_FEATURE_REDEEM_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: redeem command dry-run requires MHCAT_FEATURE_REDEEM_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: autochat config command dry-run requires MHCAT_FEATURE_AUTOCHAT_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: auto-notification config command dry-run requires MHCAT_FEATURE_AUTO_NOTIFICATION_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: anti-scam config command dry-run requires MHCAT_FEATURE_ANTI_SCAM_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: anti-scam report command dry-run requires MHCAT_FEATURE_ANTI_SCAM_REPORT_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT:-false}" = "true" ] && [ -z "${MHCAT_REPORT_WEBHOOK_URL:-}" ] && [ -z "${REPORT_WEBHOOK:-}" ]; then
  echo "refusing command sync: anti-scam report command dry-run requires MHCAT_REPORT_WEBHOOK_URL or REPORT_WEBHOOK for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_LOGGING_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: logging config command dry-run requires MHCAT_FEATURE_LOGGING_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST:-false}" = "true" ] && [ "${MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: gacha prize-list command dry-run requires MHCAT_FEATURE_GACHA_PRIZE_LIST_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW:-false}" = "true" ] && [ "${MHCAT_FEATURE_GACHA_DRAW_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: gacha draw command dry-run requires MHCAT_FEATURE_GACHA_DRAW_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE:-false}" = "true" ] && [ "${MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: gacha prize-create command dry-run requires MHCAT_FEATURE_GACHA_PRIZE_CREATE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT:-false}" = "true" ] && [ "${MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: gacha prize-edit command dry-run requires MHCAT_FEATURE_GACHA_PRIZE_EDIT_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE:-false}" = "true" ] && [ "${MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: gacha prize-delete command dry-run requires MHCAT_FEATURE_GACHA_PRIZE_DELETE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND:-false}" = "true" ] && [ "${MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: lottery disabled command dry-run requires MHCAT_FEATURE_LOTTERY_DISABLED_COMMAND_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY:-false}" = "true" ] && [ "${MHCAT_FEATURE_STATS_QUERY_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: stats query command dry-run requires MHCAT_FEATURE_STATS_QUERY_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE:-false}" = "true" ] && [ "${MHCAT_FEATURE_STATS_CREATE_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: stats create command dry-run requires MHCAT_FEATURE_STATS_CREATE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT:-false}" = "true" ] && [ "${MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: stats role-count command dry-run requires MHCAT_FEATURE_STATS_ROLE_COUNT_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE:-false}" = "true" ] && [ "${MHCAT_FEATURE_STATS_DELETE_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: stats delete command dry-run requires MHCAT_FEATURE_STATS_DELETE_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: birthday config command dry-run requires MHCAT_FEATURE_BIRTHDAY_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: announcement config command dry-run requires MHCAT_FEATURE_ANNOUNCEMENT_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND:-false}" = "true" ] && [ "${MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: announcement send command dry-run requires MHCAT_FEATURE_ANNOUNCEMENT_SEND_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: text XP config command dry-run requires MHCAT_FEATURE_TEXT_XP_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: voice XP config command dry-run requires MHCAT_FEATURE_VOICE_XP_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: XP role config commands dry-run requires MHCAT_FEATURE_XP_ROLE_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS:-false}" = "true" ] && [ "${MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: XP profile disabled commands dry-run requires MHCAT_FEATURE_XP_PROFILE_DISABLED_COMMANDS_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN:-false}" = "true" ] && [ "${MHCAT_FEATURE_XP_ADMIN_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: XP admin command dry-run requires MHCAT_FEATURE_XP_ADMIN_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: voice-room config commands dry-run requires MHCAT_FEATURE_VOICE_ROOM_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK:-false}" = "true" ] && [ "${MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: voice-room lock command dry-run requires MHCAT_FEATURE_VOICE_ROOM_LOCK_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: join-role config command dry-run requires MHCAT_FEATURE_JOIN_ROLE_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: welcome-message config command dry-run requires MHCAT_FEATURE_WELCOME_MESSAGE_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: verification config command dry-run requires MHCAT_FEATURE_VERIFICATION_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW:-false}" = "true" ] && [ "${MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: verification flow command dry-run requires MHCAT_FEATURE_VERIFICATION_FLOW_ENABLED=true for staging runtime parity" >&2
  exit 1
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG:-false}" = "true" ] && [ "${MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED:-false}" != "true" ]; then
  echo "refusing command sync: account-age config command dry-run requires MHCAT_FEATURE_ACCOUNT_AGE_CONFIG_ENABLED=true for staging runtime parity" >&2
  exit 1
fi

if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TICKETS:-false}" = "true" ]; then
  echo "staging command sync dry-run: including ticket commands for review" >&2
else
  echo "staging command sync dry-run: ticket commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_POLLS:-false}" = "true" ]; then
  echo "staging command sync dry-run: including poll commands for review" >&2
else
  echo "staging command sync dry-run: poll commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_QUERY:-false}" = "true" ]; then
  echo "staging command sync dry-run: including economy query command for review" >&2
else
  echo "staging command sync dry-run: economy query command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SIGNIN:-false}" = "true" ]; then
  echo "staging command sync dry-run: including economy sign-in commands for review" >&2
else
  echo "staging command sync dry-run: economy sign-in commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_SETTINGS:-false}" = "true" ]; then
  echo "staging command sync dry-run: including economy settings command for review" >&2
else
  echo "staging command sync dry-run: economy settings command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_ADMIN:-false}" = "true" ]; then
  echo "staging command sync dry-run: including economy coin-admin command for review" >&2
else
  echo "staging command sync dry-run: economy coin-admin command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_COIN_RANK:-false}" = "true" ]; then
  echo "staging command sync dry-run: including economy coin-rank command for review" >&2
else
  echo "staging command sync dry-run: economy coin-rank command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_RPS:-false}" = "true" ]; then
  echo "staging command sync dry-run: including economy RPS command for review" >&2
else
  echo "staging command sync dry-run: economy RPS command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ECONOMY_PROFILE:-false}" = "true" ]; then
  echo "staging command sync dry-run: including economy profile command for review" >&2
else
  echo "staging command sync dry-run: economy profile command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WORK:-false}" = "true" ]; then
  echo "staging command sync dry-run: including work command for review" >&2
else
  echo "staging command sync dry-run: work command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNINGS:-false}" = "true" ]; then
  echo "staging command sync dry-run: including warning-history command for review" >&2
else
  echo "staging command sync dry-run: warning-history command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_SETTINGS:-false}" = "true" ]; then
  echo "staging command sync dry-run: including warning-settings command for review" >&2
else
  echo "staging command sync dry-run: warning-settings command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_REMOVAL:-false}" = "true" ]; then
  echo "staging command sync dry-run: including warning-removal commands for review" >&2
else
  echo "staging command sync dry-run: warning-removal commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WARNING_ISSUE:-false}" = "true" ]; then
  echo "staging command sync dry-run: including warning issue command for review" >&2
else
  echo "staging command sync dry-run: warning issue command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_MESSAGE_CLEANUP:-false}" = "true" ]; then
  echo "staging command sync dry-run: including message cleanup command for review" >&2
else
  echo "staging command sync dry-run: message cleanup command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_DELETE_DATA:-false}" = "true" ]; then
  echo "staging command sync dry-run: including delete data command for review" >&2
else
  echo "staging command sync dry-run: delete data command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TRANSLATE:-false}" = "true" ]; then
  echo "staging command sync dry-run: including translate command for review" >&2
else
  echo "staging command sync dry-run: translate command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BALANCE_QUERY:-false}" = "true" ]; then
  echo "staging command sync dry-run: including balance query command for review" >&2
else
  echo "staging command sync dry-run: balance query command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_REDEEM:-false}" = "true" ]; then
  echo "staging command sync dry-run: including redeem command for review" >&2
else
  echo "staging command sync dry-run: redeem command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTOCHAT_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including autochat config commands for review" >&2
else
  echo "staging command sync dry-run: autochat config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_AUTO_NOTIFICATION_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including auto-notification config commands for review" >&2
else
  echo "staging command sync dry-run: auto-notification config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including anti-scam config command for review" >&2
else
  echo "staging command sync dry-run: anti-scam config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANTI_SCAM_REPORT:-false}" = "true" ]; then
  echo "staging command sync dry-run: including anti-scam report command for review" >&2
else
  echo "staging command sync dry-run: anti-scam report command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOGGING_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including logging config command for review" >&2
else
  echo "staging command sync dry-run: logging config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_LIST:-false}" = "true" ]; then
  echo "staging command sync dry-run: including gacha prize-list command for review" >&2
else
  echo "staging command sync dry-run: gacha prize-list command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_DRAW:-false}" = "true" ]; then
  echo "staging command sync dry-run: including gacha draw command for review" >&2
else
  echo "staging command sync dry-run: gacha draw command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_CREATE:-false}" = "true" ]; then
  echo "staging command sync dry-run: including gacha prize-create command for review" >&2
else
  echo "staging command sync dry-run: gacha prize-create command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_EDIT:-false}" = "true" ]; then
  echo "staging command sync dry-run: including gacha prize-edit command for review" >&2
else
  echo "staging command sync dry-run: gacha prize-edit command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_GACHA_PRIZE_DELETE:-false}" = "true" ]; then
  echo "staging command sync dry-run: including gacha prize-delete command for review" >&2
else
  echo "staging command sync dry-run: gacha prize-delete command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_LOTTERY_DISABLED_COMMAND:-false}" = "true" ]; then
  echo "staging command sync dry-run: including lottery disabled command for review" >&2
else
  echo "staging command sync dry-run: lottery disabled command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_QUERY:-false}" = "true" ]; then
  echo "staging command sync dry-run: including stats query command for review" >&2
else
  echo "staging command sync dry-run: stats query command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_CREATE:-false}" = "true" ]; then
  echo "staging command sync dry-run: including stats create command for review" >&2
else
  echo "staging command sync dry-run: stats create command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_ROLE_COUNT:-false}" = "true" ]; then
  echo "staging command sync dry-run: including stats role-count command for review" >&2
else
  echo "staging command sync dry-run: stats role-count command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_STATS_DELETE:-false}" = "true" ]; then
  echo "staging command sync dry-run: including stats delete command for review" >&2
else
  echo "staging command sync dry-run: stats delete command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_BIRTHDAY_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including birthday config command for review" >&2
else
  echo "staging command sync dry-run: birthday config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including announcement config command for review" >&2
else
  echo "staging command sync dry-run: announcement config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ANNOUNCEMENT_SEND:-false}" = "true" ]; then
  echo "staging command sync dry-run: including announcement send command for review" >&2
else
  echo "staging command sync dry-run: announcement send command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_TEXT_XP_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including text XP config commands for review" >&2
else
  echo "staging command sync dry-run: text XP config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_XP_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including voice XP config commands for review" >&2
else
  echo "staging command sync dry-run: voice XP config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_ROLE_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including XP role config commands for review" >&2
else
  echo "staging command sync dry-run: XP role config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_PROFILE_DISABLED_COMMANDS:-false}" = "true" ]; then
  echo "staging command sync dry-run: including XP profile disabled commands for review" >&2
else
  echo "staging command sync dry-run: XP profile disabled commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_XP_ADMIN:-false}" = "true" ]; then
  echo "staging command sync dry-run: including XP admin command for review" >&2
else
  echo "staging command sync dry-run: XP admin command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including voice-room config commands for review" >&2
else
  echo "staging command sync dry-run: voice-room config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VOICE_ROOM_LOCK:-false}" = "true" ]; then
  echo "staging command sync dry-run: including voice-room lock command for review" >&2
else
  echo "staging command sync dry-run: voice-room lock command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_JOIN_ROLE_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including join-role config commands for review" >&2
else
  echo "staging command sync dry-run: join-role config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_WELCOME_MESSAGE_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including welcome-message config commands for review" >&2
else
  echo "staging command sync dry-run: welcome-message config commands are excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including verification config command for review" >&2
else
  echo "staging command sync dry-run: verification config command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_VERIFICATION_FLOW:-false}" = "true" ]; then
  echo "staging command sync dry-run: including verification flow command for review" >&2
else
  echo "staging command sync dry-run: verification flow command is excluded" >&2
fi
if [ "${MHCAT_COMMAND_SYNC_INCLUDE_ACCOUNT_AGE_CONFIG:-false}" = "true" ]; then
  echo "staging command sync dry-run: including account-age config command for review" >&2
else
  echo "staging command sync dry-run: account-age config command is excluded" >&2
fi

MHCAT_COMMAND_SYNC_SCOPE=guild \
MHCAT_COMMAND_SYNC_GUILD_ID="$MHCAT_STAGING_GUILD_ID" \
MHCAT_COMMAND_SYNC_DRY_RUN=true \
MHCAT_COMMAND_SYNC_ALLOW_DELETE=false \
MHCAT_COMMAND_SYNC_ALLOW_BULK_OVERWRITE=false \
go run ./cmd/mhcat-command-sync \
  --scope guild \
  --guild-id "$MHCAT_STAGING_GUILD_ID" \
  --dry-run
