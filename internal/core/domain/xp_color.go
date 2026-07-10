package domain

import (
	"regexp"
	"strings"

	"golang.org/x/image/colornames"
)

// These patterns mirror validate-color 2.2.4, pinned by the legacy package lock.
const (
	xpColorWhitespace              = `[\x09-\x0d\x20\x{00a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]`
	xpColorSpaceNoneOrMore         = `(` + xpColorWhitespace + `{0,5})`
	xpColorSpaceOneOrMore          = `(` + xpColorWhitespace + `{1,5})`
	xpColorDigitNoneOrMore         = `([\d]{0,5})`
	xpColorDigitOneOrMore          = `([\d]{1,5})`
	xpColorOptionalCommaOrSpace    = `((` + xpColorSpaceNoneOrMore + `,` + xpColorSpaceNoneOrMore + `)|(` + xpColorSpaceOneOrMore + `))`
	xpColorOptionalDecimals        = `((\.` + xpColorDigitOneOrMore + `)?)`
	xpColorAnyNumber               = `(` + xpColorDigitNoneOrMore + xpColorOptionalDecimals + `)`
	xpColorNumberWithinHundred     = `(([0-9]|[1-9][0-9])?` + xpColorOptionalDecimals + `|100)`
	xpColorNumberWithinThreeSixty  = `(([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-9][0-9]|3[0-5][0-9])` + xpColorOptionalDecimals + `|360)`
	xpColorAnyPercentage           = `(` + xpColorAnyNumber + `%)`
	xpColorPercentageWithinHundred = `(` + xpColorNumberWithinHundred + `(%)?)`
	xpColorHundredPercent          = `(([0-9]|[1-9][0-9]|100)%)`
	xpColorAlphaPercentage         = `(((` + xpColorHundredPercent + `))|(0?` + xpColorOptionalDecimals + `)|1))?`
	xpColorAlphaRequired           = `(` + xpColorHundredPercent + `|(0?` + xpColorOptionalDecimals + `)|1)`
	xpColorEndingWithAlpha         = xpColorSpaceNoneOrMore + `\)?)` + `(` + xpColorSpaceNoneOrMore + `(\/?)` + xpColorSpaceOneOrMore + xpColorAlphaPercentage + xpColorSpaceNoneOrMore + `\)`
	xpColorDegrees                 = `(-?` + xpColorNumberWithinThreeSixty + `(deg)?)`
)

var (
	xpColorHexPattern = regexp.MustCompile(`(?i)^(?:#(?:[\da-f]{3}){1,2}|#(?:[\da-f]{4}){1,2})$`)
	xpColorRGBPattern = regexp.MustCompile(xpColorRGBRegex())
	xpColorHSLPattern = regexp.MustCompile(xpColorHSLRegex())
	xpColorHWBPattern = regexp.MustCompile(xpColorHWBRegex())
	xpColorLABPattern = regexp.MustCompile(xpColorLABRegex())
	xpColorLCHPattern = regexp.MustCompile(xpColorLCHRegex())
)

func ValidLegacyXPColor(value string) bool {
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	if _, ok := colornames.Map[lower]; ok {
		return true
	}
	switch lower {
	case "rebeccapurple", "currentcolor", "inherit", "transparent":
		return true
	}
	for _, pattern := range []*regexp.Regexp{
		xpColorHexPattern,
		xpColorRGBPattern,
		xpColorHSLPattern,
		xpColorHWBPattern,
		xpColorLABPattern,
		xpColorLCHPattern,
	} {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

func xpColorRGBRegex() string {
	letter := xpColorSpaceNoneOrMore + xpColorDigitOneOrMore + `%?` + xpColorSpaceNoneOrMore + `,?`
	gap := `((` + xpColorSpaceNoneOrMore + `,?` + xpColorSpaceNoneOrMore + `)|(` + xpColorSpaceOneOrMore + `))`
	alpha := `(\/?` + xpColorSpaceNoneOrMore + `(0?\.?` + xpColorDigitOneOrMore + `%?` + xpColorSpaceNoneOrMore + `)?|1|0)`
	return `^(rgb)a?\(` + letter + gap + letter + gap + letter + gap + `(` + alpha + `)?\)$`
}

func xpColorHSLRegex() string {
	grad := `(([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-9][0-9]|3[0-9][0-9]|400)grad)`
	rad := `((([0-5])?\.` + xpColorDigitOneOrMore + `|6\.([0-9]|1[0-9]|2[0-8])|[0-6])rad)`
	turn := `((0?` + xpColorOptionalDecimals + `|1)turn)`
	return `^(hsl)a?\((` +
		`(` + xpColorSpaceNoneOrMore + `(` + xpColorDegrees + `|` + grad + `|` + rad + `|` + turn + `)` + xpColorOptionalCommaOrSpace + `)` +
		`(` + xpColorSpaceNoneOrMore + `(0|` + xpColorHundredPercent + `)` + xpColorOptionalCommaOrSpace + `)` +
		`(` + xpColorSpaceNoneOrMore + `(0|` + xpColorHundredPercent + `)` + xpColorSpaceNoneOrMore + `\)?)` +
		`(` + xpColorSpaceNoneOrMore + `(\/?|,?)` + xpColorSpaceNoneOrMore + `(((` + xpColorHundredPercent + `))|(0?` + xpColorOptionalDecimals + `)|1))?)\)$`
}

func xpColorHWBRegex() string {
	return `^(hwb\(` + xpColorSpaceNoneOrMore + xpColorDegrees + xpColorSpaceOneOrMore + `)` +
		`((0|` + xpColorHundredPercent + `)` + xpColorSpaceOneOrMore + `)` +
		`((0|` + xpColorHundredPercent + `)` + xpColorEndingWithAlpha + `$`
}

func xpColorLABRegex() string {
	parameter := `(-?(([0-9]|[1-9][0-9]|1[0-5][0-9])` + xpColorOptionalDecimals + `?|160))`
	return `^(lab\(` + xpColorSpaceNoneOrMore + xpColorAnyPercentage + xpColorSpaceOneOrMore + parameter +
		xpColorSpaceOneOrMore + parameter + xpColorEndingWithAlpha + `$`
}

func xpColorLCHRegex() string {
	hue := `((` + xpColorDegrees + `)|(0|` + xpColorAlphaRequired + `)|` + xpColorNumberWithinThreeSixty + `)`
	alpha := `(\/` + xpColorSpaceNoneOrMore + xpColorAlphaRequired + `)`
	properties := `(` + xpColorSpaceNoneOrMore + xpColorPercentageWithinHundred + xpColorSpaceOneOrMore + xpColorAnyNumber +
		xpColorSpaceOneOrMore + hue + xpColorSpaceNoneOrMore + `(` + alpha + `)?)`
	return `^lch\(` + properties + `\)$`
}
