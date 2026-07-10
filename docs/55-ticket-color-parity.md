# Ticket Color Parity Follow-up

Status: corrected after executable dependency audit. Legacy source was not modified.

## Legacy Evidence

`events/modal.js` first validates through `validate-color@2.2.4` and then passes the same raw string to `discord.js@14.25.1`:

```txt
validateHTMLColorName(color) || validateHTMLColor(color)
new EmbedBuilder().setColor(color)
```

Both stages must accept the value. The validator alone accepts more forms than Discord's builder, so the earlier conclusion that all CSS names and 3-digit hex values worked was incorrect.

Successful legacy inputs are:

- exactly `#RRGGBB`, case-insensitive for the six hexadecimal digits;
- exact case-sensitive names present in both the HTML-name validator and Discord's `Colors` object, such as `Red`, `Aqua`, `DarkGreen`, and `LightGrey`.

Discord names resolve to Discord palette values. For example, `Red` is `0xED4245`, not CSS red `0xFF0000`. Values such as `red`, `#0f0`, `rebeccapurple`, `DarkSlateGray`, and `rgb(1,2,3)` pass one stage at most and do not produce a legacy panel.

## Change

The Go parser now models the successful two-stage intersection. It does not trim input, accepts only 6-digit hash-prefixed hex, and uses the exact Discord palette for the shared case-sensitive names.

## Tests

Table tests cover successful hex and named inputs, Discord palette values, case sensitivity, whitespace, short/alpha hex rejection, CSS-only names, and CSS functional forms.
