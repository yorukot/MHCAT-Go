package economy

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type signCalendarView struct {
	Year       int
	Month      time.Month
	Username   string
	StatusText string
	Calendar   domain.SignCalendar
}

func renderSignPNG(view signCalendarView) ([]byte, error) {
	if view.Year < 1 || view.Month < time.January || view.Month > time.December {
		return nil, domain.ErrInvalidSignIn
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 1000, 707))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 22, G: 27, B: 42, A: 255}}, image.Point{}, draw.Src)
	fillRect(canvas, image.Rect(30, 28, 970, 675), color.RGBA{R: 38, G: 44, B: 64, A: 255})
	fillRect(canvas, image.Rect(49, 147, 951, 649), color.RGBA{R: 18, G: 22, B: 34, A: 255})
	drawText(canvas, 100, 89, fmt.Sprintf("%04d/%02d", view.Year, int(view.Month)), color.RGBA{R: 0, G: 255, B: 255, A: 255}, 4)
	drawText(canvas, 680, 89, view.Username, color.RGBA{R: 235, G: 239, B: 255, A: 255}, 3)
	for _, line := range []struct {
		x1, y1, x2, y2 int
	}{
		{49, 197, 951, 197}, {49, 272, 951, 272}, {49, 347, 951, 347}, {49, 422, 951, 422}, {49, 497, 951, 497}, {49, 572, 951, 572},
		{177, 147, 177, 649}, {305, 147, 305, 649}, {433, 147, 433, 649}, {561, 147, 561, 649}, {689, 147, 689, 649}, {817, 147, 817, 649},
	} {
		drawLine(canvas, line.x1, line.y1, line.x2, line.y2, color.White)
	}
	for i, label := range []string{"Sun.", "Mon.", "Tue.", "Wed.", "Thu.", "Fri.", "Sat."} {
		drawText(canvas, 69+i*128, 185, label, color.RGBA{R: 255, G: 211, B: 6, A: 255}, 2)
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
			drawText(canvas, x, y, strconv.Itoa(day), textColor, 3)
			if view.Calendar.HasDay(yearKey, monthKey, strconv.Itoa(day)) {
				drawCheck(canvas, x+60, y-45)
			}
		}
	}
	drawText(canvas, 120, 690, view.StatusText, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 2)
	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
		for _, path := range []string{
			"../MHCAT/fonts/TaipeiSansTCBeta-Regular.ttf",
			"MHCAT/fonts/TaipeiSansTCBeta-Regular.ttf",
			"./fonts/TaipeiSansTCBeta-Regular.ttf",
		} {
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
