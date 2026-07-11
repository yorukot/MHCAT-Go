# Slash Command UI Parity Audit

Generated from the current local legacy `MHCAT/slashCommands` tree and the current Go command definitions.
This is a static slash-command metadata audit for names, descriptions, options, and option flags; handler behavior remains covered by feature-specific tests and docs.
Default permission discoverability and runtime permission responses are outside this generic audit; announcement and anti-scam permission behavior is locked in their [announcement](76-announcement.md) and [anti-scam](77-anti-scam.md) contracts.
Rerun with `go run ./tools/parity-audit --legacy-root ../MHCAT --format markdown`.

## Summary

- Legacy slash command files: 74
- Legacy unique command names: 74
- Current Go command definitions: 74
- Matching command definitions: 74
- Implemented definitions needing UI review: 0
- Legacy commands without Go definitions: 0
- Go definitions without a legacy command name: 0
- Legacy parse warning/error files: 0

## Matching Definitions

| Command | Legacy file | Status | Findings |
| --- | --- | --- | --- |
| `automatic-notification` | `slashCommands/自動通知/cron_set.js` | matching-definition | none |
| `coin-related-settings` | `slashCommands/代幣系統/coin_cange.js` | matching-definition | none |
| `help` | `slashCommands/實用工具/help.js` | matching-definition | none |
| `info` | `slashCommands/實用工具/info.js` | matching-definition | none |
| `my-profile` | `slashCommands/代幣系統/user-info.js` | matching-definition | none |
| `ping` | `slashCommands/實用工具/ping.js` | matching-definition | none |
| `set-log-channel` | `slashCommands/管理系統/create_logging.js` | matching-definition | none |
| `上鎖頻道` | `slashCommands/語音包廂/lock_channel.js` | matching-definition | none |
| `代幣商店` | `slashCommands/代幣系統/ghp_shop.js` | matching-definition | none |
| `代幣增加` | `slashCommands/代幣系統/addcoin.js` | matching-definition | none |
| `代幣排行榜` | `slashCommands/代幣系統/coin_rank.js` | matching-definition | none |
| `代幣查詢` | `slashCommands/代幣系統/coin.js` | matching-definition | none |
| `代幣遊戲` | `slashCommands/代幣系統/game.js` | matching-definition | none |
| `代幣重製` | `slashCommands/代幣系統/coin_rest.js` | matching-definition | none |
| `兌換` | `slashCommands/管理系統/get_something.js` | matching-definition | none |
| `公告發送` | `slashCommands/公告系統/announcement.js` | matching-definition | none |
| `公告頻道設置` | `slashCommands/公告系統/announcement_set_channel.js` | matching-definition | none |
| `刪除訊息` | `slashCommands/管理系統/clear.js` | matching-definition | none |
| `刪除資料` | `slashCommands/管理系統/delete_data.js` | matching-definition | none |
| `剪刀石頭布` | `slashCommands/代幣系統/rock_paper_scissors.js` | matching-definition | none |
| `加入訊息設置` | `slashCommands/加入設置/join_messag.js` | matching-definition | none |
| `加入身份組刪除` | `slashCommands/加入設置/join_role_delete.js` | matching-definition | none |
| `加入身份組設置` | `slashCommands/加入設置/join_role.js` | matching-definition | none |
| `帳號需創建時數` | `slashCommands/群組防護/create_hours.js` | matching-definition | none |
| `打工系統` | `slashCommands/打工系統/work_set.js` | matching-definition | none |
| `扭蛋` | `slashCommands/扭蛋系統/gashapon.js` | matching-definition | none |
| `扭蛋獎品編輯` | `slashCommands/扭蛋系統/giftadd copy.js` | matching-definition | none |
| `扭蛋獎池刪除` | `slashCommands/扭蛋系統/gift_delete.js` | matching-definition | none |
| `扭蛋獎池增加` | `slashCommands/扭蛋系統/giftadd.js` | matching-definition | none |
| `扭蛋獎池查詢` | `slashCommands/扭蛋系統/giftlist.js` | matching-definition | none |
| `投票創建` | `slashCommands/管理系統/poll.js` | matching-definition | none |
| `抽獎設置` | `slashCommands/抽獎系統/lotter_create.js` | matching-definition | none |
| `查看餘額` | `slashCommands/管理系統/check_price.js` | matching-definition | none |
| `生日系統` | `slashCommands/生日系統/birthday.js` | parity-audited | See [birthday parity contract](78-birthday.md) |
| `私人頻道刪除` | `slashCommands/私人頻道/ticket_delete.js` | matching-definition | none |
| `私人頻道設置` | `slashCommands/私人頻道/ticket.js` | matching-definition | none |
| `簽到` | `slashCommands/代幣系統/sing.js` | matching-definition | none |
| `簽到列表` | `slashCommands/代幣系統/sing_list.js` | matching-definition | none |
| `統計系統刪除` | `slashCommands/統計系統/number_delete.js` | matching-definition | none |
| `統計系統創建` | `slashCommands/統計系統/number_create.js` | matching-definition | none |
| `統計系統查詢` | `slashCommands/統計系統/number.js` | matching-definition | none |
| `統計身分組人數` | `slashCommands/統計系統/role_create.js` | matching-definition | none |
| `經驗值改變` | `slashCommands/經驗系統/xp_add.js` | matching-definition | none |
| `經驗值重製` | `slashCommands/經驗系統/reset_xp.js` | matching-definition | none |
| `翻譯` | `slashCommands/實用工具/translate.js` | matching-definition | none |
| `聊天排行榜` | `slashCommands/經驗系統/text_rank.js` | matching-definition | none |
| `聊天經驗` | `slashCommands/經驗系統/text_xp.js` | matching-definition | none |
| `聊天經驗刪除` | `slashCommands/經驗系統/text_set_delete.js` | matching-definition | none |
| `聊天經驗設定` | `slashCommands/經驗系統/text_set.js` | matching-definition | none |
| `聊天經驗身分組設定` | `slashCommands/經驗系統/text_leave_role.js` | matching-definition | none |
| `自動聊天頻道` | `slashCommands/實用工具/chat.js` | matching-definition | none |
| `自動聊天頻道刪除` | `slashCommands/實用工具/chat_delete.js` | matching-definition | none |
| `自動通知列表` | `slashCommands/自動通知/cron_list.js` | matching-definition | none |
| `自動通知刪除` | `slashCommands/自動通知/cron_delete.js` | matching-definition | none |
| `詐騙網址回報` | `slashCommands/群組防護/report_web.js` | matching-definition | none |
| `語音包廂刪除` | `slashCommands/語音包廂/voice_channel_delete.js` | matching-definition | none |
| `語音包廂設置` | `slashCommands/語音包廂/voice_channel.js` | matching-definition | none |
| `語音排行榜` | `slashCommands/經驗系統/voice_rank.js` | matching-definition | none |
| `語音經驗` | `slashCommands/經驗系統/voice_xp.js` | matching-definition | none |
| `語音經驗刪除` | `slashCommands/經驗系統/voice_set_delete.js` | matching-definition | none |
| `語音經驗設定` | `slashCommands/經驗系統/voice_set.js` | matching-definition | none |
| `語音經驗身分組設定` | `slashCommands/經驗系統/voice_leavel_role.js` | matching-definition | none |
| `警告` | `slashCommands/警告系統/warn.js` | matching-definition | none |
| `警告全部清除` | `slashCommands/警告系統/remove-all-warnings.js` | matching-definition | none |
| `警告清除` | `slashCommands/警告系統/remove-warn.js` | matching-definition | none |
| `警告紀錄` | `slashCommands/警告系統/warnings.js` | matching-definition | none |
| `警告設定` | `slashCommands/警告系統/erros_set.js` | matching-definition | none |
| `退出訊息設置` | `slashCommands/加入設置/leave_message.js` | matching-definition | none |
| `選取身分組-按鈕` | `slashCommands/管理系統/releadd.js` | matching-definition | none |
| `選取身分組-表情符號` | `slashCommands/管理系統/role.js` | matching-definition | none |
| `選取身分組刪除-表情符號` | `slashCommands/管理系統/role_delete.js` | matching-definition | none |
| `防詐騙網址` | `slashCommands/群組防護/not_a_goodweb.js` | matching-definition | none |
| `驗證` | `slashCommands/加入設置/verification.js` | matching-definition | none |
| `驗證設置` | `slashCommands/加入設置/verification_set.js` | matching-definition | none |

## Implemented Definitions Needing UI Review

None.

## Legacy Commands Without Go Definitions

None.

## Go Definitions Without Legacy Command Names

None.
