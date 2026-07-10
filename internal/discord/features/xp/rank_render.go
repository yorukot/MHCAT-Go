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
	ViewerRankText string
	Title          string
	Entries        []rankCanvasEntry
}

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
	fillRankRect(canvas, image.Rect(33, 10, 103, 80), color.RGBA{R: 88, G: 101, B: 242, A: 255})
	drawRankText(canvas, 45, 55, "M", color.RGBA{R: 255, G: 255, B: 255, A: 255}, 3)
	guildName := truncateLegacyRankText(view.GuildName)
	if guildName == "" {
		guildName = "MHCAT"
	}
	createdAt := view.GuildCreatedAt
	if createdAt.IsZero() {
		createdAt = time.Unix(0, 0).UTC()
	}
	drawRankText(canvas, 115, 50, guildName, color.RGBA{R: 211, G: 211, B: 211, A: 255}, 3)
	drawRankText(canvas, 118, 74, view.Title, color.RGBA{R: 168, G: 168, B: 168, A: 255}, 2)
	drawRankText(canvas, 710, 70, view.ViewerRankText, color.RGBA{R: 230, G: 235, B: 255, A: 255}, 3)
	drawRankText(canvas, 790, 70, createdAt.Format("2006/01/02"), color.RGBA{R: 230, G: 235, B: 255, A: 255}, 2)
}

func drawRankRows(canvas *image.RGBA, view rankCanvasView) {
	for i, entry := range view.Entries {
		xOffset := 0
		row := i
		if i >= 5 {
			xOffset = 484
			row = i - 5
		}
		rankY := 146 + row*74
		nameY := 131 + row*74
		xpY := 153 + row*74
		rankScale := 4
		if entry.Rank > 99 && entry.Rank < 1000 {
			rankScale = 3
		}
		if entry.Rank > 999 {
			rankScale = 2
		}
		drawRankText(canvas, 55+xOffset, rankY, strconv.Itoa(entry.Rank), color.RGBA{R: 255, G: 255, B: 255, A: 255}, rankScale)
		drawRankText(canvas, 121+xOffset, nameY, truncateLegacyRankText(entry.DisplayName), color.RGBA{R: 255, G: 255, B: 255, A: 255}, 2)
		drawRankText(canvas, 137+xOffset, xpY, coreservice.LegacyRankAmount(entry.TotalXP), color.RGBA{R: 226, G: 230, B: 240, A: 255}, 2)
	}
}

func fillRankRect(img *image.RGBA, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func drawRankText(img *image.RGBA, x, y int, text string, c color.RGBA, scale int) {
	if face := rankFontFace(float64(scale) * 10); face != nil {
		defer face.Close()
		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(c),
			Face: face,
			Dot:  fixed.P(x, y),
		}
		drawer.DrawString(text)
		return
	}
	cursor := x
	for _, r := range text {
		drawRankRune(img, cursor, y, r, c, scale)
		cursor += 6*scale + scale
	}
}

var rankFont struct {
	once sync.Once
	font *opentype.Font
}

func rankFontFace(size float64) font.Face {
	rankFont.once.Do(func() {
		for _, path := range append(rankAssetCandidates("fonts/TaipeiSansTCBeta-Regular.ttf"), "./fonts/TaipeiSansTCBeta-Regular.ttf") {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			parsed, err := opentype.Parse(data)
			if err == nil {
				rankFont.font = parsed
				return
			}
		}
	})
	if rankFont.font == nil {
		return nil
	}
	face, err := opentype.NewFace(rankFont.font, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}
	return face
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
