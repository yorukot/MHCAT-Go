package economy

import (
	"image"
	"image/color"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var profileXPFont struct {
	once sync.Once
	font *opentype.Font
}

func drawProfileXPTextCentered(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	face := profileXPFontFace(float64(size))
	if face == nil {
		drawCoinRankCenteredNumericText(img, x, y, text, c, size)
		return
	}
	defer face.Close()
	x -= font.MeasureString(face, text).Ceil() / 2
	drawer := &font.Drawer{Dst: img, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, y)}
	drawer.DrawString(text)
}

func profileXPFontFace(size float64) font.Face {
	profileXPFont.once.Do(func() {
		for _, candidate := range legacyAssetCandidates("fonts/Oswald-Regular.ttf") {
			data, err := os.ReadFile(candidate)
			if err != nil {
				continue
			}
			parsed, err := opentype.Parse(data)
			if err == nil {
				profileXPFont.font = parsed
				return
			}
		}
	})
	if profileXPFont.font == nil {
		return nil
	}
	face, err := opentype.NewFace(profileXPFont.font, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}
	return face
}
