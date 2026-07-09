# Ticket Legacy Panel Submit Compatibility

Status: implemented behind the existing ticket feature/runtime gates. Legacy source was not modified.

## Scope

This follow-up closes the ticket `nal` modal gap found during welcome-message review:

- `私人頻道設置` now checks Manage Messages at runtime and returns the legacy permission-denied embed.
- Existing ticket config returns the legacy `__**錯誤**__` duplicate-config embed and does not overwrite the stored config.
- Versioned Go setup modal submit now defers, validates the legacy fields, saves config after validation, sends the panel to the current channel, and edits the deferred reply with the legacy success embed.
- Legacy live modal submit `nal` + `ticketcolor` now routes to `ticket/panel_submit` and sends the same panel/success response without needing versioned payload fields.

## Preserved Legacy UI

- Modal title: `私人頻道系統!`
- Modal fields: `ticketcolor`, `tickettitle`, `ticketcontent`
- Panel button custom ID: `tic`
- Panel button label: `🎫 點我創建客服頻道!`
- Invalid color title: `你傳送的並不是顏色(色碼)`
- Color validation accepts 3/6-digit hex and CSS/HTML named colors used by the legacy `validate-color` path.
- Success title: `<a:green_tick:994529015652163614> | 成功創建私人頻道`

## Intentional Difference

Legacy saves the `tickets` config before the modal is submitted. The Go path keeps the documented bug fix: versioned setup writes config only after the modal validates. This prevents abandoned or invalid setup modals from leaving a live ticket config without a panel.

## Tests

Added coverage for:

- setup permission-denied legacy embed;
- setup duplicate-config legacy embed and no overwrite;
- valid versioned modal submit sends a channel panel and edits success;
- invalid versioned modal submit edits the legacy color error without saving;
- legacy `nal` panel submit sends a channel panel and edits success;
- legacy `nal` invalid color path;
- CSS/HTML named color parity for representative legacy-accepted names;
- module route registration for `ticket/panel_submit`.
