package economy

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type signCalendarView struct {
	Year       int
	Month      time.Month
	Username   string
	AvatarData []byte
	StatusText string
	Calendar   domain.SignCalendar
}

var signBackgroundCache struct {
	once  sync.Once
	image image.Image
}

func renderSignPNG(view signCalendarView) ([]byte, error) {
	if view.Year < 1 || view.Month < time.January || view.Month > time.December {
		return nil, domain.ErrInvalidSignIn
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 1000, 707))
	drawSignBackground(canvas)
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{A: 128}}, image.Point{}, draw.Over)
	drawSignAsset(canvas, "asset/mhcat_white.png", image.Pt(20, 35))
	drawSignAvatar(canvas, view.AvatarData)
	drawCoinRankNumericText(canvas, 100, 89, fmt.Sprintf("%04d/%02d", view.Year, int(view.Month)), color.RGBA{R: 0, G: 255, B: 255, A: 255}, 40)
	drawSignRightAlignedText(canvas, 880, 89, view.Username, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 45)
	for _, line := range []struct {
		x1, y1, x2, y2 int
	}{
		{49, 197, 951, 197}, {49, 272, 951, 272}, {49, 347, 951, 347}, {49, 422, 951, 422}, {49, 497, 951, 497}, {49, 572, 951, 572},
		{177, 147, 177, 649}, {305, 147, 305, 649}, {433, 147, 433, 649}, {561, 147, 561, 649}, {689, 147, 689, 649}, {817, 147, 817, 649},
	} {
		drawLine(canvas, line.x1, line.y1, line.x2, line.y2, color.White)
	}
	for i, label := range []string{"Sun.", "Mon.", "Tue.", "Wed.", "Thu.", "Fri.", "Sat."} {
		drawCoinRankNumericText(canvas, 69+i*128, 185, label, color.RGBA{R: 255, G: 211, B: 6, A: 255}, 40)
	}
	yearKey := fmt.Sprintf("%04d", view.Year)
	monthKey := fmt.Sprintf("%02d", int(view.Month))
	weeks := monthGrid(view.Year, view.Month)
	for row, week := range weeks {
		for column, day := range week {
			if day == 0 {
				continue
			}
			textColor := color.RGBA{R: 168, G: 255, B: 36, A: 255}
			if column == 0 || column == 6 {
				textColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
			}
			x := 55 + column*128
			y := 252 + row*75
			drawCoinRankNumericText(canvas, x, y, strconv.Itoa(day), textColor, 45)
			if view.Calendar.HasDay(yearKey, monthKey, strconv.Itoa(day)) {
				drawSignAsset(canvas, "asset/verify_icon.png", image.Pt(115+column*128, 202+row*75))
			}
		}
	}
	drawCoinRankAlignedText(canvas, 500, 690, view.StatusText, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 30, true, coinRankLanguageFont)
	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawSignAvatar(canvas *image.RGBA, data []byte) {
	avatar := decodeProfileImage(data)
	if avatar == nil {
		avatar = loadSignAsset("asset/yellow_discord.png")
	}
	if avatar == nil {
		return
	}
	large := image.NewRGBA(image.Rect(0, 0, 128, 128))
	xdraw.CatmullRom.Scale(large, large.Bounds(), avatar, avatar.Bounds(), draw.Over, nil)
	rounded := image.NewRGBA(large.Bounds())
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			if insideRoundedRect(x, y, 128, 128, 40) {
				rounded.Set(x, y, large.At(x, y))
			}
		}
	}
	small := image.NewRGBA(image.Rect(0, 0, 80, 80))
	xdraw.CatmullRom.Scale(small, small.Bounds(), rounded, rounded.Bounds(), draw.Over, nil)
	draw.Draw(canvas, image.Rect(900, 35, 980, 115), small, image.Point{}, draw.Over)
}

func drawSignBackground(canvas *image.RGBA) {
	signBackgroundCache.once.Do(func() {
		background := loadSignAsset("asset/background.png")
		if background == nil {
			return
		}
		rgba := image.NewRGBA(image.Rect(0, 0, 1000, 707))
		draw.Draw(rgba, rgba.Bounds(), background, background.Bounds().Min, draw.Src)
		signBackgroundCache.image = gaussianBlurSignBackground(rgba)
	})
	if signBackgroundCache.image != nil {
		draw.Draw(canvas, canvas.Bounds(), signBackgroundCache.image, image.Point{}, draw.Src)
		return
	}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 22, G: 27, B: 42, A: 255}}, image.Point{}, draw.Src)
}

func gaussianBlurSignBackground(source *image.RGBA) *image.RGBA {
	const radius = 5
	const sigma = 5.0
	weights := make([]float64, radius*2+1)
	total := 0.0
	for offset := -radius; offset <= radius; offset++ {
		weight := 1 / (math.Sqrt(2*math.Pi) * sigma) * math.Exp(-float64(offset*offset)/(2*sigma*sigma))
		weights[offset+radius] = weight
		total += weight
	}
	for index := range weights {
		weights[index] /= total
	}
	bounds := source.Bounds()
	result := image.NewRGBA(bounds)
	draw.Draw(result, bounds, source, bounds.Min, draw.Src)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var red, green, blue, sum float64
			for offset := -radius; offset <= radius; offset++ {
				sampleX := x + offset
				if sampleX < bounds.Min.X || sampleX >= bounds.Max.X {
					continue
				}
				pixel := result.RGBAAt(sampleX, y)
				weight := weights[offset+radius]
				red += float64(pixel.R) * weight
				green += float64(pixel.G) * weight
				blue += float64(pixel.B) * weight
				sum += weight
			}
			result.SetRGBA(x, y, color.RGBA{R: uint8(red / sum), G: uint8(green / sum), B: uint8(blue / sum), A: result.RGBAAt(x, y).A})
		}
	}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			var red, green, blue, sum float64
			for offset := -radius; offset <= radius; offset++ {
				sampleY := y + offset
				if sampleY < bounds.Min.Y || sampleY >= bounds.Max.Y {
					continue
				}
				pixel := result.RGBAAt(x, sampleY)
				weight := weights[offset+radius]
				red += float64(pixel.R) * weight
				green += float64(pixel.G) * weight
				blue += float64(pixel.B) * weight
				sum += weight
			}
			result.SetRGBA(x, y, color.RGBA{R: uint8(red / sum), G: uint8(green / sum), B: uint8(blue / sum), A: result.RGBAAt(x, y).A})
		}
	}
	return result
}

func loadSignAsset(relative string) image.Image {
	for _, candidate := range legacyAssetCandidates(relative) {
		file, err := os.Open(candidate)
		if err != nil {
			continue
		}
		asset, _, err := image.Decode(file)
		_ = file.Close()
		if err == nil {
			return asset
		}
	}
	return nil
}

func drawSignAsset(canvas *image.RGBA, relative string, point image.Point) {
	asset := loadSignAsset(relative)
	if asset == nil {
		return
	}
	destination := image.Rectangle{Min: point, Max: point.Add(asset.Bounds().Size())}
	draw.Draw(canvas, destination, asset, asset.Bounds().Min, draw.Over)
}

func drawSignRightAlignedText(canvas *image.RGBA, x int, y int, text string, c color.RGBA, size int) {
	face := coinRankFontFace(coinRankLanguageFont, float64(size))
	if face == nil {
		drawCoinRankText(canvas, x-coinRankFallbackTextWidth(text, max(1, (size+5)/10)), y, text, c, size)
		return
	}
	defer face.Close()
	drawCoinRankText(canvas, x-font.MeasureString(face, text).Ceil(), y, text, c, size)
}

func monthGrid(year int, month time.Month) [][7]int {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	var weeks [][7]int
	week := [7]int{}
	column := int(first.Weekday())
	for day := 1; day <= last; day++ {
		week[column] = day
		column++
		if column == 7 {
			weeks = append(weeks, week)
			week = [7]int{}
			column = 0
		}
	}
	if column != 0 {
		weeks = append(weeks, week)
	}
	return weeks
}

func fillRect(img *image.RGBA, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	if x1 == x2 {
		fillRect(img, image.Rect(x1, y1, x1+2, y2+2), c)
		return
	}
	fillRect(img, image.Rect(x1, y1, x2+2, y1+2), c)
}

func drawCheck(img *image.RGBA, x, y int) {
	fillRect(img, image.Rect(x, y+20, x+10, y+30), color.RGBA{R: 0, G: 255, B: 180, A: 255})
	fillRect(img, image.Rect(x+10, y+30, x+20, y+40), color.RGBA{R: 0, G: 255, B: 180, A: 255})
	fillRect(img, image.Rect(x+20, y+10, x+30, y+40), color.RGBA{R: 0, G: 255, B: 180, A: 255})
	fillRect(img, image.Rect(x+30, y, x+40, y+20), color.RGBA{R: 0, G: 255, B: 180, A: 255})
}

func drawText(img *image.RGBA, x, y int, text string, c color.RGBA, scale int) {
	if face := signFontFace(float64(scale) * 10); face != nil {
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
		drawRune(img, cursor, y, r, c, scale)
		cursor += 6*scale + scale
	}
}

var signFont struct {
	once sync.Once
	font *opentype.Font
}

func signFontFace(size float64) font.Face {
	signFont.once.Do(func() {
		for _, path := range append(legacyAssetCandidates("fonts/TaipeiSansTCBeta-Regular.ttf"), "./fonts/TaipeiSansTCBeta-Regular.ttf") {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			parsed, err := opentype.Parse(data)
			if err == nil {
				signFont.font = parsed
				return
			}
		}
	})
	if signFont.font == nil {
		return nil
	}
	face, err := opentype.NewFace(signFont.font, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}
	return face
}

func drawRune(img *image.RGBA, x, y int, r rune, c color.RGBA, scale int) {
	pattern, ok := font5x7[r]
	if !ok {
		pattern = font5x7['?']
	}
	for row, bits := range pattern {
		for col := 0; col < 5; col++ {
			if bits&(1<<(4-col)) == 0 {
				continue
			}
			fillRect(img, image.Rect(x+col*scale, y+row*scale, x+(col+1)*scale, y+(row+1)*scale), c)
		}
	}
}

var font5x7 = map[rune][7]byte{
	' ': {0, 0, 0, 0, 0, 0, 0},
	'?': {0x0E, 0x11, 0x01, 0x02, 0x04, 0, 0x04},
	'.': {0, 0, 0, 0, 0, 0x0C, 0x0C},
	'/': {0x01, 0x02, 0x04, 0x08, 0x10, 0, 0},
	'-': {0, 0, 0, 0x1F, 0, 0, 0},
	'0': {0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E},
	'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
	'2': {0x0E, 0x11, 0x01, 0x02, 0x04, 0x08, 0x1F},
	'3': {0x1F, 0x02, 0x04, 0x02, 0x01, 0x11, 0x0E},
	'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
	'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
	'6': {0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E},
	'7': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08},
	'8': {0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E},
	'9': {0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C},
	'A': {0x0E, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
	'D': {0x1E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x1E},
	'F': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x10},
	'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
	'S': {0x0F, 0x10, 0x10, 0x0E, 0x01, 0x01, 0x1E},
	'T': {0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04},
	'W': {0x11, 0x11, 0x11, 0x15, 0x15, 0x15, 0x0A},
	'a': {0, 0x0E, 0x01, 0x0F, 0x11, 0x13, 0x0D},
	'd': {0x01, 0x01, 0x0D, 0x13, 0x11, 0x13, 0x0D},
	'e': {0, 0x0E, 0x11, 0x1F, 0x10, 0x11, 0x0E},
	'h': {0x10, 0x10, 0x16, 0x19, 0x11, 0x11, 0x11},
	'i': {0x04, 0, 0x0C, 0x04, 0x04, 0x04, 0x0E},
	'n': {0, 0, 0x16, 0x19, 0x11, 0x11, 0x11},
	'o': {0, 0, 0x0E, 0x11, 0x11, 0x11, 0x0E},
	'r': {0, 0, 0x16, 0x19, 0x10, 0x10, 0x10},
	't': {0x08, 0x08, 0x1C, 0x08, 0x08, 0x09, 0x06},
	'u': {0, 0, 0x11, 0x11, 0x11, 0x13, 0x0D},
}
