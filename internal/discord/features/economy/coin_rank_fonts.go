package economy

import (
	"image"
	"image/color"
	"os"
	"sync"
	"unicode/utf16"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type coinRankFontFamily uint8

const (
	coinRankLanguageFont coinRankFontFamily = iota
	coinRankNumericFont
)

func drawCoinRankText(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	drawCoinRankAlignedText(img, x, y, text, c, size, false, coinRankLanguageFont)
}

func drawCoinRankNumericText(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	drawCoinRankAlignedText(img, x, y, text, c, size, false, coinRankFontFamilyForNumericText(text))
}

func drawCoinRankCenteredNumericText(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	drawCoinRankAlignedText(img, x, y, text, c, size, true, coinRankFontFamilyForNumericText(text))
}

func drawCoinRankAlignedText(img *image.RGBA, x, y int, text string, c color.RGBA, size int, centered bool, family coinRankFontFamily) {
	if face := coinRankFontFace(family, float64(size)); face != nil {
		defer face.Close()
		if centered {
			x -= font.MeasureString(face, text).Ceil() / 2
		}
		drawer := &font.Drawer{Dst: img, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, y)}
		drawer.DrawString(text)
		return
	}
	scale := max(1, (size+5)/10)
	if centered {
		x -= coinRankFallbackTextWidth(text, scale) / 2
	}
	cursor := x
	for _, r := range text {
		drawRune(img, cursor, y-7*scale, r, c, scale)
		cursor += 7 * scale
	}
}

func coinRankFontFamilyForNumericText(text string) coinRankFontFamily {
	for _, r := range text {
		if r > 0x7f {
			return coinRankLanguageFont
		}
	}
	return coinRankNumericFont
}

func coinRankFallbackTextWidth(text string, scale int) int {
	count := 0
	for range text {
		count++
	}
	if count == 0 {
		return 0
	}
	return (count*7 - 2) * scale
}

var coinRankFonts = [2]struct {
	once  sync.Once
	fonts []*opentype.Font
}{}

func coinRankFontFace(family coinRankFontFamily, size float64) font.Face {
	cache := &coinRankFonts[family]
	cache.once.Do(func() {
		for _, relative := range coinRankFontCandidates(family) {
			for _, candidate := range legacyAssetCandidates(relative) {
				data, err := os.ReadFile(candidate)
				if err != nil {
					continue
				}
				parsed, err := opentype.Parse(data)
				if err == nil {
					cache.fonts = append(cache.fonts, parsed)
					break
				}
			}
		}
	})
	if len(cache.fonts) == 0 && family == coinRankNumericFont {
		return coinRankFontFace(coinRankLanguageFont, size)
	}
	if len(cache.fonts) == 0 {
		return nil
	}
	faces := make([]font.Face, 0, len(cache.fonts))
	for _, parsed := range cache.fonts {
		face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
		if err == nil {
			faces = append(faces, face)
		}
	}
	if len(faces) == 0 && family == coinRankNumericFont {
		return coinRankFontFace(coinRankLanguageFont, size)
	}
	if len(faces) == 0 {
		return nil
	}
	if len(faces) == 1 {
		return faces[0]
	}
	return &coinRankFallbackFace{faces: faces}
}

func coinRankFontCandidates(family coinRankFontFamily) []string {
	candidates := []string{
		"fonts/language/TC.otf",
		"fonts/language/SC.otf",
		"fonts/language/JP.otf",
		"fonts/language/HK.otf",
		"fonts/language/NotoSans.ttf",
		"fonts/language/Bengali.ttf",
		"fonts/language/Arabic.ttf",
		"fonts/language/emoji.ttf",
		"fonts/TaipeiSansTCBeta-Regular.ttf",
	}
	if family == coinRankNumericFont {
		return append([]string{"fonts/Comic-Sans-MS-copy-5-.ttf"}, candidates...)
	}
	return candidates
}

type coinRankFallbackFace struct{ faces []font.Face }

func (f *coinRankFallbackFace) Close() error {
	var firstErr error
	for _, face := range f.faces {
		if err := face.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (f *coinRankFallbackFace) Glyph(dot fixed.Point26_6, r rune) (image.Rectangle, image.Image, image.Point, fixed.Int26_6, bool) {
	for _, face := range f.faces {
		if bounds, mask, maskPoint, advance, ok := face.Glyph(dot, r); ok {
			return bounds, mask, maskPoint, advance, true
		}
	}
	return image.Rectangle{}, nil, image.Point{}, 0, false
}

func (f *coinRankFallbackFace) GlyphBounds(r rune) (fixed.Rectangle26_6, fixed.Int26_6, bool) {
	for _, face := range f.faces {
		if bounds, advance, ok := face.GlyphBounds(r); ok {
			return bounds, advance, true
		}
	}
	return fixed.Rectangle26_6{}, 0, false
}

func (f *coinRankFallbackFace) GlyphAdvance(r rune) (fixed.Int26_6, bool) {
	for _, face := range f.faces {
		if advance, ok := face.GlyphAdvance(r); ok {
			return advance, true
		}
	}
	return 0, false
}

func (f *coinRankFallbackFace) Kern(r0 rune, r1 rune) fixed.Int26_6 {
	for _, face := range f.faces {
		if _, ok := face.GlyphAdvance(r0); !ok {
			continue
		}
		if _, ok := face.GlyphAdvance(r1); ok {
			return face.Kern(r0, r1)
		}
	}
	return 0
}

func (f *coinRankFallbackFace) Metrics() font.Metrics {
	if len(f.faces) == 0 {
		return font.Metrics{}
	}
	return f.faces[0].Metrics()
}

func legacyCoinRankSlotPosition(slot int) (int, int) {
	if slot >= 5 {
		return 484, slot - 5
	}
	return 0, slot
}

func legacyCoinRankNumber(page int, slot int) int { return page*10 + slot + 1 }

func legacyCoinRankNumberFontSize(page int) int {
	if page > 99 && page < 1000 {
		return 30
	}
	if page > 999 {
		return 25
	}
	return 40
}

func truncateLegacyCoinRankText(value string) string {
	units := utf16.Encode([]rune(value))
	if legacyCoinRankTextWidth(units) <= 34 {
		return value
	}
	limit := 33
	prefixEnd := min(limit, len(units))
	if legacyCoinRankTextWidth(units[:prefixEnd]) > 34 {
		limit = 16
	}
	return string(utf16.Decode(units[:min(limit, len(units))]))
}

func legacyCoinRankTextWidth(units []uint16) int {
	width := 0
	for _, unit := range units {
		width++
		if unit > 0xff {
			width++
		}
	}
	return width
}

func insideLegacyCoinRankIcon(x int, y int, width int, height int, radius int) bool {
	if x >= radius && x < width-radius || y >= radius && y < height-radius {
		return true
	}
	cx := radius
	if x >= width-radius {
		cx = width - radius - 1
	}
	cy := radius
	if y >= height-radius {
		cy = height - radius - 1
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= radius*radius
}
