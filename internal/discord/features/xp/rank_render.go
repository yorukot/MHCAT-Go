package xp

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"unicode/utf16"

	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/xp"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type rankCanvasEntry struct {
	Rank        int
	DisplayName string
	TotalXP     int64
}

type rankCanvasView struct {
	GuildName      string
	GuildCreatedAt time.Time
	GuildIconData  []byte
	ViewerRankText string
	Title          string
	Page           int
	Entries        []rankCanvasEntry
}

type rankFontFamily uint8

const (
	rankLanguageFont rankFontFamily = iota
	rankNumericFont
)

func renderRankPNG(view rankCanvasView) ([]byte, error) {
	canvas := image.NewRGBA(image.Rect(0, 0, 1000, 500))
	if !drawRankBackground(canvas) {
		drawGeneratedRankBackground(canvas)
	}
	drawRankHeader(canvas, view)
	drawRankRows(canvas, view)

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawRankBackground(canvas *image.RGBA) bool {
	for _, path := range rankAssetCandidates("asset/rank_background.png") {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		img, _, err := image.Decode(file)
		_ = file.Close()
		if err != nil {
			continue
		}
		draw.Draw(canvas, canvas.Bounds(), img, img.Bounds().Min, draw.Src)
		return true
	}
	return false
}

func drawGeneratedRankBackground(canvas *image.RGBA) {
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 18, G: 24, B: 32, A: 255}}, image.Point{}, draw.Src)
	fillRankRect(canvas, image.Rect(24, 18, 976, 88), color.RGBA{R: 47, G: 56, B: 69, A: 255})
	fillRankRect(canvas, image.Rect(24, 102, 482, 478), color.RGBA{R: 30, G: 37, B: 49, A: 255})
	fillRankRect(canvas, image.Rect(518, 102, 976, 478), color.RGBA{R: 30, G: 37, B: 49, A: 255})
	for y := 176; y < 478; y += 74 {
		fillRankRect(canvas, image.Rect(24, y, 482, y+2), color.RGBA{R: 65, G: 73, B: 91, A: 255})
		fillRankRect(canvas, image.Rect(518, y, 976, y+2), color.RGBA{R: 65, G: 73, B: 91, A: 255})
	}
}

func drawRankHeader(canvas *image.RGBA, view rankCanvasView) {
	drawRankGuildIcon(canvas, view.GuildIconData)
	guildName := truncateLegacyRankText(view.GuildName)
	if guildName == "" {
		guildName = "MHCAT"
	}
	createdAt := view.GuildCreatedAt
	if createdAt.IsZero() {
		createdAt = time.Unix(0, 0).UTC()
	}
	headerColor := color.RGBA{R: 211, G: 211, B: 211, A: 255}
	drawRankText(canvas, 115, 50, guildName, headerColor, 37)
	drawRankText(canvas, 118, 74, view.Title, color.RGBA{R: 168, G: 168, B: 168, A: 255}, 20)
	drawRankCenteredNumericText(canvas, 710, 70, view.ViewerRankText, headerColor, 30)
	drawRankNumericText(canvas, 790, 70, createdAt.Format("2006/01/02"), headerColor, 30)
}

func drawRankGuildIcon(canvas *image.RGBA, data []byte) {
	icon := decodeRankImage(data)
	if icon == nil {
		icon = legacyRankFallbackIcon()
	}
	if icon == nil {
		fillRankRect(canvas, image.Rect(33, 10, 103, 80), color.RGBA{R: 88, G: 101, B: 242, A: 255})
		drawRankText(canvas, 45, 55, "M", color.RGBA{R: 255, G: 255, B: 255, A: 255}, 30)
		return
	}
	large := image.NewRGBA(image.Rect(0, 0, 128, 128))
	xdraw.CatmullRom.Scale(large, large.Bounds(), icon, icon.Bounds(), draw.Over, nil)
	rounded := image.NewRGBA(large.Bounds())
	for y := 0; y < large.Bounds().Dy(); y++ {
		for x := 0; x < large.Bounds().Dx(); x++ {
			if insideLegacyRankIcon(x, y, 128, 128, 40) {
				rounded.Set(x, y, large.At(x, y))
			}
		}
	}
	small := image.NewRGBA(image.Rect(0, 0, 70, 70))
	xdraw.CatmullRom.Scale(small, small.Bounds(), rounded, rounded.Bounds(), draw.Over, nil)
	draw.Draw(canvas, image.Rect(33, 10, 103, 80), small, image.Point{}, draw.Over)
}

func decodeRankImage(data []byte) image.Image {
	if len(data) == 0 {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return img
}

func legacyRankFallbackIcon() image.Image {
	for _, path := range rankAssetCandidates("asset/blue_discord.png") {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		img, _, err := image.Decode(file)
		_ = file.Close()
		if err == nil {
			return img
		}
	}
	return nil
}

func insideLegacyRankIcon(x int, y int, width int, height int, radius int) bool {
	if x >= radius && x < width-radius {
		return true
	}
	if y >= radius && y < height-radius {
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
	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy <= radius*radius
}

func drawRankRows(canvas *image.RGBA, view rankCanvasView) {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	rankSize := legacyRankNumberFontSize(view.Page)
	for slot := 0; slot < coreservice.RankPageSize; slot++ {
		xOffset, row := legacyRankSlotPosition(slot)
		drawRankCenteredNumericText(canvas, 73+xOffset, 146+row*74, strconv.Itoa(legacyRankNumber(view.Page, slot)), white, rankSize)
	}
	for i, entry := range view.Entries {
		xOffset, row := legacyRankSlotPosition(i)
		nameY := 131 + row*74
		xpY := 153 + row*74
		drawRankText(canvas, 121+xOffset, nameY, truncateLegacyRankText(entry.DisplayName), white, 25)
		drawRankNumericText(canvas, 137+xOffset, xpY, coreservice.LegacyRankAmount(entry.TotalXP), white, 15)
	}
}

func legacyRankSlotPosition(slot int) (int, int) {
	if slot >= coreservice.RankPageSize/2 {
		return 484, slot - coreservice.RankPageSize/2
	}
	return 0, slot
}

func legacyRankNumber(page int, slot int) int {
	return page*coreservice.RankPageSize + slot + 1
}

func legacyRankNumberFontSize(page int) int {
	if page > 99 && page < 1000 {
		return 30
	}
	if page > 999 {
		return 25
	}
	return 40
}

func fillRankRect(img *image.RGBA, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func drawRankText(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	drawRankAlignedText(img, x, y, text, c, size, false, rankLanguageFont)
}

func drawRankNumericText(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	drawRankAlignedText(img, x, y, text, c, size, false, rankFontFamilyForNumericText(text))
}

func drawRankCenteredNumericText(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	drawRankAlignedText(img, x, y, text, c, size, true, rankFontFamilyForNumericText(text))
}

func drawRankAlignedText(img *image.RGBA, x, y int, text string, c color.RGBA, size int, centered bool, family rankFontFamily) {
	if face := rankFontFace(family, float64(size)); face != nil {
		defer face.Close()
		if centered {
			x -= font.MeasureString(face, text).Ceil() / 2
		}
		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(c),
			Face: face,
			Dot:  fixed.P(x, y),
		}
		drawer.DrawString(text)
		return
	}
	scale := max(1, (size+5)/10)
	if centered {
		x -= rankFallbackTextWidth(text, scale) / 2
	}
	cursor := x
	for _, r := range text {
		drawRankRune(img, cursor, y-7*scale, r, c, scale)
		cursor += 6*scale + scale
	}
}

func rankFontFamilyForNumericText(text string) rankFontFamily {
	for _, r := range text {
		if r > 0x7f {
			return rankLanguageFont
		}
	}
	return rankNumericFont
}

func rankFallbackTextWidth(text string, scale int) int {
	count := 0
	for range text {
		count++
	}
	if count == 0 {
		return 0
	}
	return (count*7 - 2) * scale
}

var rankFonts = [2]struct {
	once sync.Once
	font *opentype.Font
}{}

func rankFontFace(family rankFontFamily, size float64) font.Face {
	cache := &rankFonts[family]
	cache.once.Do(func() {
		for _, relative := range rankFontCandidates(family) {
			for _, path := range rankAssetCandidates(relative) {
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				parsed, err := opentype.Parse(data)
				if err == nil {
					cache.font = parsed
					return
				}
			}
		}
	})
	if cache.font == nil && family == rankNumericFont {
		return rankFontFace(rankLanguageFont, size)
	}
	if cache.font == nil {
		return nil
	}
	face, err := opentype.NewFace(cache.font, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}
	return face
}

func rankFontCandidates(family rankFontFamily) []string {
	if family == rankNumericFont {
		return []string{
			"fonts/Comic-Sans-MS-copy-5-.ttf",
			"fonts/language/TC.otf",
			"fonts/TaipeiSansTCBeta-Regular.ttf",
		}
	}
	return []string{
		"fonts/language/TC.otf",
		"fonts/TaipeiSansTCBeta-Regular.ttf",
	}
}

func drawRankRune(img *image.RGBA, x, y int, r rune, c color.RGBA, scale int) {
	pattern, ok := rankFont5x7[r]
	if !ok {
		pattern = rankFont5x7['?']
	}
	for row, bits := range pattern {
		for col := 0; col < 5; col++ {
			if bits&(1<<(4-col)) == 0 {
				continue
			}
			fillRankRect(img, image.Rect(x+col*scale, y+row*scale, x+(col+1)*scale, y+(row+1)*scale), c)
		}
	}
}

func rankAssetCandidates(relative string) []string {
	return []string{
		relative,
		filepath.Join("MHCAT", relative),
		filepath.Join("..", "MHCAT", relative),
		filepath.Join("..", "..", "..", "..", "..", "MHCAT", relative),
	}
}

func truncateLegacyRankText(value string) string {
	units := utf16.Encode([]rune(value))
	if legacyRankTextWidth(units) <= 34 {
		return value
	}
	limit := 33
	prefixEnd := min(limit, len(units))
	if legacyRankTextWidth(units[:prefixEnd]) > 34 {
		limit = 16
	}
	return string(utf16.Decode(units[:min(limit, len(units))]))
}

func legacyRankTextWidth(units []uint16) int {
	width := 0
	for _, unit := range units {
		width++
		if unit > 0xff {
			width++
		}
	}
	return width
}

var rankFont5x7 = map[rune][7]byte{
	' ': {0, 0, 0, 0, 0, 0, 0},
	'?': {0x1E, 0x11, 0x01, 0x02, 0x04, 0, 0x04},
	'#': {0x0A, 0x1F, 0x0A, 0x0A, 0x1F, 0x0A, 0x0A},
	'/': {0x01, 0x02, 0x04, 0x04, 0x08, 0x10, 0x00},
	'-': {0, 0, 0, 0x1F, 0, 0, 0},
	'.': {0, 0, 0, 0, 0, 0x0C, 0x0C},
	'0': {0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E},
	'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'2': {0x0E, 0x11, 0x01, 0x02, 0x04, 0x08, 0x1F},
	'3': {0x1E, 0x01, 0x01, 0x0E, 0x01, 0x01, 0x1E},
	'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
	'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
	'6': {0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E},
	'7': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08},
	'8': {0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E},
	'9': {0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C},
	'A': {0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'G': {0x0E, 0x11, 0x10, 0x17, 0x11, 0x11, 0x0E},
	'K': {0x11, 0x12, 0x14, 0x18, 0x14, 0x12, 0x11},
	'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
}
