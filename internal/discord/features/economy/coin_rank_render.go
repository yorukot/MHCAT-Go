package economy

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"time"

	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	xdraw "golang.org/x/image/draw"
)

type coinRankCanvasEntry struct {
	Rank        int
	DisplayName string
	Coins       string
}

type coinRankCanvasView struct {
	GuildName      string
	GuildCreatedAt time.Time
	GuildIconData  []byte
	ViewerRankText string
	SubtitleY      int
	Page           int
	Entries        []coinRankCanvasEntry
}

func renderCoinRankPNG(view coinRankCanvasView) ([]byte, error) {
	canvas := image.NewRGBA(image.Rect(0, 0, 1000, 500))
	if !drawCoinRankBackground(canvas) {
		drawGeneratedCoinRankBackground(canvas)
	}
	drawCoinRankHeader(canvas, view)
	drawCoinRankRows(canvas, view)

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawCoinRankBackground(canvas *image.RGBA) bool {
	for _, path := range legacyAssetCandidates("asset/coin_rank_background.png") {
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

func drawGeneratedCoinRankBackground(canvas *image.RGBA) {
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 20, G: 24, B: 34, A: 255}}, image.Point{}, draw.Src)
	fillRect(canvas, image.Rect(24, 18, 976, 88), color.RGBA{R: 45, G: 51, B: 69, A: 255})
	fillRect(canvas, image.Rect(24, 102, 482, 478), color.RGBA{R: 31, G: 36, B: 50, A: 255})
	fillRect(canvas, image.Rect(518, 102, 976, 478), color.RGBA{R: 31, G: 36, B: 50, A: 255})
	for y := 176; y < 478; y += 74 {
		fillRect(canvas, image.Rect(24, y, 482, y+2), color.RGBA{R: 65, G: 72, B: 91, A: 255})
		fillRect(canvas, image.Rect(518, y, 976, y+2), color.RGBA{R: 65, G: 72, B: 91, A: 255})
	}
}

func drawCoinRankHeader(canvas *image.RGBA, view coinRankCanvasView) {
	drawCoinRankGuildIcon(canvas, view.GuildIconData)
	guildName := truncateLegacyCoinRankText(view.GuildName)
	if guildName == "" {
		guildName = "MHCAT"
	}
	createdAt := view.GuildCreatedAt
	if createdAt.IsZero() {
		createdAt = time.Unix(0, 0).UTC()
	}
	headerColor := color.RGBA{R: 211, G: 211, B: 211, A: 255}
	drawCoinRankText(canvas, 115, 50, guildName, headerColor, 37)
	subtitleY := view.SubtitleY
	if subtitleY == 0 {
		subtitleY = 74
	}
	drawCoinRankText(canvas, 118, subtitleY, "代幣排行榜", color.RGBA{R: 168, G: 168, B: 168, A: 255}, 20)
	drawCoinRankCenteredNumericText(canvas, 710, 70, view.ViewerRankText, headerColor, 30)
	drawCoinRankNumericText(canvas, 790, 70, createdAt.Format("2006/01/02"), headerColor, 30)
}

func drawCoinRankGuildIcon(canvas *image.RGBA, data []byte) {
	icon := decodeProfileImage(data)
	if icon == nil {
		icon = legacyCoinRankFallbackIcon()
	}
	if icon == nil {
		fillRect(canvas, image.Rect(33, 10, 103, 80), color.RGBA{R: 88, G: 101, B: 242, A: 255})
		return
	}
	large := image.NewRGBA(image.Rect(0, 0, 128, 128))
	xdraw.CatmullRom.Scale(large, large.Bounds(), icon, icon.Bounds(), draw.Over, nil)
	rounded := image.NewRGBA(large.Bounds())
	for y := 0; y < large.Bounds().Dy(); y++ {
		for x := 0; x < large.Bounds().Dx(); x++ {
			if insideLegacyCoinRankIcon(x, y, 128, 128, 40) {
				rounded.Set(x, y, large.At(x, y))
			}
		}
	}
	small := image.NewRGBA(image.Rect(0, 0, 70, 70))
	xdraw.CatmullRom.Scale(small, small.Bounds(), rounded, rounded.Bounds(), draw.Over, nil)
	draw.Draw(canvas, image.Rect(33, 10, 103, 80), small, image.Point{}, draw.Over)
}

func legacyCoinRankFallbackIcon() image.Image {
	for _, path := range legacyAssetCandidates("asset/blue_discord.png") {
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

func drawCoinRankRows(canvas *image.RGBA, view coinRankCanvasView) {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	rankSize := legacyCoinRankNumberFontSize(view.Page)
	for slot := 0; slot < coreeconomy.CoinRankPageSize; slot++ {
		xOffset, row := legacyCoinRankSlotPosition(slot)
		drawCoinRankCenteredNumericText(canvas, 73+xOffset, 146+row*74, strconv.Itoa(legacyCoinRankNumber(view.Page, slot)), white, rankSize)
	}
	for i, entry := range view.Entries {
		xOffset, row := legacyCoinRankSlotPosition(i)
		nameY := 131 + row*74
		coinsY := 153 + row*74
		drawCoinRankText(canvas, 121+xOffset, nameY, truncateLegacyCoinRankText(entry.DisplayName), white, 25)
		drawCoinRankNumericText(canvas, 137+xOffset, coinsY, coreeconomy.LegacyCoinRankAmountText(entry.Coins), white, 15)
	}
}

func legacyAssetCandidates(relative string) []string {
	return []string{
		relative,
		filepath.Join("MHCAT", relative),
		filepath.Join("..", "MHCAT", relative),
		filepath.Join("..", "..", "..", "..", "..", "MHCAT", relative),
	}
}

func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
