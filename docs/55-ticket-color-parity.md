# Ticket Color Parity Follow-up

Status: implemented. Legacy source was not modified.

## Legacy Evidence

`events/modal.js` validates ticket panel colors through:

```txt
validateHTMLColorName(color) || validateHTMLColor(color)
```

The previous Go ticket slice accepted 3/6-digit hex plus a small hand-written list of eight color names. That rejected common legacy-accepted names such as `aqua`, `rebeccapurple`, and `darkslategray`.

## Change

The Go ticket color parser now includes the standard CSS/HTML named color table and remains case-insensitive. It still supports 3-digit and 6-digit hex values.

## Tests

Added table tests for:

- legacy basic name `red`;
- names outside the previous small list: `aqua`, `rebeccapurple`, `darkslategray`;
- case-insensitive names;
- `grey` alias names;
- 3-digit, 6-digit, and uppercase hex;
- empty and unknown values rejected.

## Remaining Work

The legacy `validate-color` package is not installed in the local repository, so exact behavior for non-name CSS forms was not tested locally. The legacy code calls only `validateHTMLColorName` and `validateHTMLColor`, not RGB/HSL validators, so this slice intentionally covers named colors and hex values only.
