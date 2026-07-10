# XP Rank Parity Audit

Status: parity-audited behind explicit runtime and command-sync gates.

## Legacy References

- `MHCAT/slashCommands/經驗系統/text_rank.js`
- `MHCAT/slashCommands/經驗系統/voice_rank.js`
- `MHCAT/events/rank.js`
- `MHCAT/models/text_xp.js`
- `MHCAT/models/voice_xp.js`

## Scope And Gates

This read-only slice implements `/聊天排行榜` and `/語音排行榜`, including their legacy pagination buttons and PNG renderer. Runtime requires:

```bash
MHCAT_FEATURE_XP_RANK_ENABLED=true
```

Staging command sync also requires `MHCAT_COMMAND_SYNC_INCLUDE_XP_RANK=true`. Config validation, staging preflight, and scripts reject an enabled command-sync flag without the runtime flag. No gateway or privileged intent is required.

The slice reads `text_xps` and `voice_xps`. It does not write XP, create indexes, enable XP accrual, enable profile cards, change reward roles, or award coins.

## Preserved Interaction Contract

- Both slash commands reply publicly with the orange `#FF5809` loading embed, author `正在努力為您尋找資料!`, legacy loading GIF, footer `MHCAT 帶給你最好的discord體驗!`, and the invoking user's PNG avatar.
- A missing custom avatar uses `https://i.imgur.com/B91C90T.png`, matching the legacy `avatarURL({extension: 'png'})` fallback.
- The final edit clears the loading embed and attaches a 1000x500 PNG named `user-info.png`.
- Text and voice cards use titles `聊天經驗排行榜` and `語音經驗排行榜` respectively.
- A component whose original viewer is no longer a resolvable guild member returns the ephemeral red embed `<a:Discord_AnimatedNo:1015989839809757295> | 找不到資料!請於幾分鐘後重試!`.
- Slash usage is recorded by the global usage middleware when usage tracking is enabled. Component presses do not increment slash usage.

## Preserved Ranking Contract

- Mongo is queried by `{guild}` with no database sort. Legacy iterates that result backwards before applying its stable descending total-XP sort; Go preserves that tie order.
- The displayed sort total remains `100 + xp + sum(trunc(y * (y / 3)) * 100)` for `y` from `level - 1` through zero.
- Viewer rank intentionally uses the legacy's different comparison totals. Candidate rows add `100` for every loop iteration including `y = 0`; the viewer threshold adds it only while `y > 0`. Rank is the count of candidate totals greater than or equal to that threshold.
- Legacy obtains the viewer through a separate `findOne`. With duplicate viewer rows, Go therefore uses the first source-order profile for viewer-rank math while still showing every duplicate in the leaderboard.
- Tied totals retain reverse source order. No user-ID tie breaker is added.
- Amounts retain the legacy `K`, `M`, and `G` formatter, one decimal place, stripped `.0`, and JavaScript `toFixed(1)` rounding behavior.
- Cached Discord user tags remain `username#discriminator` for legacy discriminator accounts and `username` for migrated discriminator-`0` accounts. Missing users display `找不到該名使用者`.

## Preserved Pagination Contract

- Pages contain ten rows, arranged as five rows in each canvas column.
- The first button row retains the exact `-10`, `-1`, page label, `+1`, and `+10` IDs, emojis, success/secondary styles, and disabled conditions.
- The disabled page-label custom ID is `text_rank` for both commands. The voice command's `text_rank` label ID is an intentional legacy bug.
- The second row retains disabled `text_rank1`, `text_rank2`, `text_rank4`, and `text_rank5` spacer buttons around the active target-viewer button.
- The target page remains `trunc(viewerRank / 10)`, so viewer rank 10 targets page index 1 rather than page index 0.
- Legacy compares `沒有資料!` against `沒有資料` when deciding whether to include the second row. The comparison never matches, so Go also emits that row for viewers without an XP profile, including its active `{NaN}` target ID.
- An empty leaderboard renders `1/0`, disables all four navigation buttons, includes both button rows, and still paints rank numbers 1 through 10 with blank names and amounts.
- An ordinary nonnegative out-of-range page remains on that requested page, renders its ten rank numbers with blank entries, enables valid backward navigation, and disables forward navigation.

## Preserved Canvas Contract

- The renderer uses `asset/rank_background.png`; if the guild has no icon, it uses `asset/blue_discord.png`.
- Guild icons are requested as PNG, masked at 128x128 with the legacy radius, then scaled into the 70x70 header slot.
- Rank numbers, names, XP amounts, header text, viewer rank, and guild creation date retain the legacy coordinates, alignment, colors, and font sizes.
- Numeric fields lead with `Comic Sans MS`. Text fields use the legacy `TC`, `SC`, `JP`, `HK`, Noto, Bengali, Arabic, and emoji fallback sequence.
- Guild names and user tags preserve the legacy UTF-16 truncation rule: characters above `0xFF` count as width two, with a 33-code-unit prefix or a 16-code-unit wide-text fallback.
- Every page paints ten rank numbers even when fewer or no profiles exist. Rank-number font size remains based on the page index, not the displayed rank.

## Asset And Deployment Requirements

For the legacy appearance, the bot working tree must make these files discoverable through the documented `MHCAT` sibling layout or equivalent relative paths:

- `asset/rank_background.png`
- `asset/blue_discord.png`
- `fonts/Comic-Sans-MS-copy-5-.ttf`
- `fonts/language/{TC,SC,JP,HK}.otf`
- `fonts/language/{NotoSans,Bengali,Arabic,emoji}.ttf`
- `fonts/TaipeiSansTCBeta-Regular.ttf`

If assets are absent or invalid, Go deliberately renders a generated background, placeholder icon, and built-in bitmap text instead of leaving the loading reply unresolved.

## Intentional Go Differences

- The typed custom-ID parser accepts only exact 17-20 digit viewer IDs and bounded nonnegative integer pages. It rejects negative, fractional, `{NaN}`, suffix-injected, and otherwise malformed IDs that legacy's broad `includes()` plus `Number()` path attempted to render. Consequently, clicking the no-profile `{NaN}` target is rejected instead of producing a corrupted NaN page.
- Render-safe page bounds prevent integer overflow from crafted stale component IDs. Normal out-of-range pages remain supported.
- Legacy checks the original viewer only in `guild.members.cache`; Go checks DiscordGo state and then the guild-member REST endpoint. A current uncached member can therefore work in Go. A departed user still present only in the legacy global user cache can display differently.
- Leaderboard names in legacy come only from `client.users.cache`; Go resolves current guild members through state/REST. Both paths retain the visible tag and missing-user strings when they resolve the same user.
- Guild-icon downloads are context-bound, limited to two seconds and 2 MiB, and fall back locally on failure. Legacy's nested image promises had no equivalent bounds and could leave the loading message unchanged after rejection.
- Mongo, Discord lookup, image, and render failures are returned through Go's centralized error handling instead of being ignored inside legacy callbacks or promises.
- Go suppresses allowed mentions on all rank responses.
- Font selection, coordinates, source assets, and dimensions are preserved, but Go's image/font rasterizers and scaling filters are not pixel-identical to node-canvas. Visual staging comparison is required after font or runtime-image changes.

## Verification Coverage

Automated tests lock:

- exact command definitions and runtime/command-sync gates;
- loading and missing-viewer payloads, including avatar normalization;
- sort totals, viewer-rank quirks, reverse-source tie order, duplicate viewer selection, amount rounding, and ten-row pagination;
- exact text/voice component rows, IDs, emojis, styles, disabled states, `1/0`, voice `text_rank`, rank-10 targeting, and no-profile `{NaN}`;
- bounded parser behavior and huge out-of-range service pages without overflow;
- legacy user tags and UTF-16 truncation boundaries;
- canvas dimensions, empty-page rank slots, guild-icon masking, font-family order, multilingual fallback, and concurrent rendering under the race detector;
- read-only Mongo filters and app/config/preflight feature wiring.

## Staging Checklist

1. Keep the Node.js rank command owner stopped for the staging guild while Go owns these command names and components.
2. Verify all background, fallback-icon, and font assets above are available from the bot's production working directory.
3. Enable the runtime and command-sync flags, run `go run ./cmd/mhcat-staging-preflight`, and review command-sync dry-run before apply.
4. Compare both commands' loading embeds, final filenames, titles, guild icon, guild date, multilingual names, totals, and both button rows against legacy screenshots.
5. Test zero, one, ten, eleven, and more than twenty profiles, including tied totals and a user at displayed rank 10.
6. Test an invoking viewer with no XP profile; verify `1/0` where applicable and the preserved active `{NaN}` target row, then verify pressing that malformed target is safely rejected.
7. Press `-10`, `-1`, `+1`, `+10`, and target-viewer controls on text and voice pages, including an ordinary stale out-of-range page.
8. Remove the original viewer from the staging guild and verify a stale component returns the exact ephemeral missing-user embed.
9. Confirm Mongo profiler/audit output contains reads only and no indexes or XP writes are created by this feature.
10. Re-run image comparison after changing the Go image library, font files, legacy assets, or bot working directory.
