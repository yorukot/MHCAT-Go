# Poll Wave B: Result Chart and Export

## Scope

Implemented the second poll parity slice:

- `see_result` now returns the legacy-style result embed with `attachment://file.jpg`;
- result responses attach a generated `file.jpg` pie chart;
- result responses attach `discord.txt` with legacy per-vote row text;
- owner-menu `poll_excel_result` now returns `poll_info.xlsx`;
- anonymous poll Excel export still returns `該投票為匿名，無法查看投票資訊!`;
- member tags are resolved through a Discord member port when available and fall back to `使用者已退出伺服器!`.

## Legacy UI Preserved

Visible output follows `events/poll.js`:

- result embed title: `<:poll:1023968837965709312> | <question>`;
- field name format: `<choice>(共<n>人 \`<percent>\`%)`;
- empty option value: `<a:Discord_AnimatedNo:1015989839809757295> | 還沒有人投給這個選項`;
- anonymous option value: `該投票為匿名，無法查看誰有進行投票`;
- large result value: `由於人數過多，無法顯示所有人`;
- Excel success content: `<:sheets:1023972957330100324> | **以下是該投票的excel表格!**`;
- Excel filename: `poll_info.xlsx`.

## Intentional Internal Changes

- The chart is rendered with Go standard-library image/JPEG code instead of Node `chartjs-node-canvas`.
- The Excel file is generated as a minimal OpenXML workbook with Go standard-library ZIP/XML code instead of `write-excel-file`.
- Time export uses an explicit Asia/Taipei format ending in `台北標準時間` to avoid locale-dependent host behavior.
- Member lookup fetches each member once through a port instead of repeatedly calling Discord fetches inline.

## Not Implemented Yet

- Real staging smoke for `/投票創建`, vote buttons, result export, and owner menu controls.
- Production poll command sync.
- Duplicate-key audit before any unique index on `{guild, messageid}`.

## Tests Added

- Result handler verifies `file.jpg` and `discord.txt` attachments.
- Artifact tests verify JPEG bytes, text export content, anonymous redaction, valid XLSX ZIP structure, XML escaping, and legacy Excel content.
- Owner menu tests verify `poll_info.xlsx` attachment and anonymous export denial.

## Next Step

Run staging guild smoke for the complete poll flow with both runtime and command-sync poll flags enabled. After smoke passes, move to the next domain slice, preferably read-only economy `/代幣查詢` or verification setup.
