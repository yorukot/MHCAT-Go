package economy

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strconv"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"

	_ "image/gif"
	_ "image/jpeg"
)

type profileCanvasView struct {
	DisplayName    string
	GuildName      string
	UserCreatedAt  time.Time
	MemberJoinedAt time.Time
	AvatarData     []byte
	Result         coreeconomy.ProfileResult
}

func renderProfilePNG(view profileCanvasView) ([]byte, error) {
	canvas := image.NewRGBA(image.Rect(0, 0, 1500, 750))
	if !drawProfileBackground(canvas) {
		drawGeneratedProfileBackground(canvas)
	}
	drawProfileAvatar(canvas, view.AvatarData)
	drawProfileHeader(canvas, view)
	drawProfileStats(canvas, view.Result)
	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawProfileBackground(canvas *image.RGBA) bool {
	for _, path := range legacyAssetCandidates("asset/background_profile.png") {
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

func drawGeneratedProfileBackground(canvas *image.RGBA) {
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 22, G: 27, B: 42, A: 255}}, image.Point{}, draw.Src)
	fillRect(canvas, image.Rect(32, 24, 1468, 138), color.RGBA{R: 45, G: 51, B: 69, A: 255})
	fillRect(canvas, image.Rect(45, 170, 500, 405), color.RGBA{R: 31, G: 36, B: 50, A: 255})
	fillRect(canvas, image.Rect(520, 170, 970, 405), color.RGBA{R: 31, G: 36, B: 50, A: 255})
	fillRect(canvas, image.Rect(1008, 170, 1460, 405), color.RGBA{R: 31, G: 36, B: 50, A: 255})
	fillRect(canvas, image.Rect(45, 455, 1460, 700), color.RGBA{R: 31, G: 36, B: 50, A: 255})
}

func drawProfileAvatar(canvas *image.RGBA, data []byte) {
	img := decodeProfileImage(data)
	if img == nil {
		img = legacyProfileFallbackAvatar()
	}
	if img == nil {
		fillRect(canvas, image.Rect(42, 30, 140, 128), color.RGBA{R: 255, G: 191, B: 64, A: 255})
		drawText(canvas, 70, 92, "M", color.RGBA{R: 255, G: 255, B: 255, A: 255}, 4)
		return
	}
	large := image.NewRGBA(image.Rect(0, 0, 128, 128))
	xdraw.CatmullRom.Scale(large, large.Bounds(), img, img.Bounds(), draw.Over, nil)
	rounded := image.NewRGBA(large.Bounds())
	for y := 0; y < large.Bounds().Dy(); y++ {
		for x := 0; x < large.Bounds().Dx(); x++ {
			if insideRoundedRect(x, y, 128, 128, 40) {
				rounded.Set(x, y, large.At(x, y))
			}
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, 98, 98))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), rounded, rounded.Bounds(), draw.Over, nil)
	draw.Draw(canvas, image.Rect(42, 30, 140, 128), dst, image.Point{}, draw.Over)
}

func decodeProfileImage(data []byte) image.Image {
	if len(data) == 0 {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return img
}

func legacyProfileFallbackAvatar() image.Image {
	for _, path := range legacyAssetCandidates("asset/yellow_discord.png") {
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

func drawRoundedImage(canvas *image.RGBA, img *image.RGBA, point image.Point, radius int) {
	bounds := img.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if !insideRoundedRect(x, y, bounds.Dx(), bounds.Dy(), radius) {
				continue
			}
			canvas.Set(point.X+x, point.Y+y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
}

func insideRoundedRect(x, y, width, height, radius int) bool {
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

func drawProfileHeader(canvas *image.RGBA, view profileCanvasView) {
	displayName := truncateRunes(view.DisplayName, 33)
	if displayName == "" {
		displayName = view.Result.UserID
	}
	drawText(canvas, 151, 80, displayName, color.RGBA{R: 211, G: 211, B: 211, A: 255}, 4)
	drawText(canvas, 151, 120, truncateRunes(view.GuildName, 46), color.RGBA{R: 168, G: 168, B: 168, A: 255}, 2)
	drawText(canvas, 1220, 100, legacyProfileDate(view.MemberJoinedAt), color.RGBA{R: 211, G: 211, B: 211, A: 255}, 3)
	drawText(canvas, 960, 100, legacyProfileDate(view.UserCreatedAt), color.RGBA{R: 211, G: 211, B: 211, A: 255}, 3)
}

func drawProfileStats(canvas *image.RGBA, result coreeconomy.ProfileResult) {
	drawProfileProgress(canvas, 550, 333, profileProgressWidth(result.TextXP.XP, coreeconomy.LegacyProfileXPRequired(result.TextXP.Level, false), result.TextXPFound), color.RGBA{R: 100, G: 255, B: 191, A: 255})
	drawProfileProgress(canvas, 1038, 333, profileProgressWidth(result.VoiceXP.XP, coreeconomy.LegacyProfileXPRequired(result.VoiceXP.Level, true), result.VoiceXPFound), color.RGBA{R: 234, G: 121, B: 255, A: 255})
	drawProfileTextCentered(canvas, 750, 363, profileXPProgressText(result.TextXP, result.TextXPFound, false), color.RGBA{R: 255, G: 88, B: 9, A: 255}, 3)
	drawProfileTextCentered(canvas, 1238, 363, profileXPProgressText(result.VoiceXP, result.VoiceXPFound, true), color.RGBA{R: 40, G: 255, B: 40, A: 255}, 3)

	drawProfileTextCentered(canvas, 367, 243, profileRankText(result.TextRank, result.TextXPFound, true), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 367, 306, profileRankText(result.VoiceRank, result.VoiceXPFound, false), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 367, 369, profileRankText(result.CoinRank, result.CoinFound, false), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)

	drawProfileTextCentered(canvas, 864, 243, profileXPValue(result.TextXP, result.TextXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 864, 306, profileLevelValue(result.TextXP, result.TextXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 1351, 243, profileXPValue(result.VoiceXP, result.VoiceXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 1351, 306, profileLevelValue(result.VoiceXP, result.VoiceXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)

	drawProfileTextCentered(canvas, 295, 525, profileCoinText(result), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 295, 587, result.SignStatus, color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 639, 525, profileWorkEnergyText(result), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 639, 587, profileWorkStateText(result), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 1045, 525, profileConfigIntText(result.Config.GachaCost, result.ConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 1045, 587, profileConfigIntText(result.Config.SignCoins, result.ConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 1385, 525, profileConfigIntText(result.WorkConfig.DailyEnergy, result.WorkConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 1385, 587, profileConfigIntText(result.WorkConfig.MaxEnergy, result.WorkConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
	drawProfileTextCentered(canvas, 1237, 652, profileConfigFloatText(result.Config.XPMultiple, result.ConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 4)
}

func drawProfileProgress(canvas *image.RGBA, x, y, width int, c color.RGBA) {
	if width <= 0 {
		return
	}
	fillRect(canvas, image.Rect(x, y, x+width, y+35), c)
}

func profileProgressWidth(xp int64, required int64, found bool) int {
	if !found || required <= 0 || xp <= 0 {
		return 0
	}
	percent := int(float64(xp)/float64(required)*100 + 0.5)
	if percent <= 0 {
		return 0
	}
	return 35 + percent*36/10
}

func profileXPProgressText(profile domain.XPProfile, found bool, voice bool) string {
	if !found {
		return "0/0"
	}
	return coreeconomy.LegacyProfileAmount(float64(profile.XP)) + "/" + coreeconomy.LegacyProfileAmount(float64(coreeconomy.LegacyProfileXPRequired(profile.Level, voice)))
}

func profileRankText(rank int, found bool, text bool) string {
	if !found {
		if text {
			return "#沒有資料!"
		}
		return "#沒有資料"
	}
	return "#" + strconv.Itoa(rank)
}

func profileXPValue(profile domain.XPProfile, found bool) string {
	if !found {
		return "沒有資料"
	}
	return coreeconomy.LegacyProfileAmount(float64(profile.XP))
}

func profileLevelValue(profile domain.XPProfile, found bool) string {
	if !found {
		return "沒有資料"
	}
	return strconv.FormatInt(profile.Level, 10)
}

func profileCoinText(result coreeconomy.ProfileResult) string {
	if !result.CoinFound {
		return "0"
	}
	return coreeconomy.LegacyProfileAmount(float64(result.CoinBalance.Coins))
}

func profileWorkEnergyText(result coreeconomy.ProfileResult) string {
	if !result.WorkUserFound {
		return "0"
	}
	return coreeconomy.LegacyProfileAmount(float64(result.WorkUser.Energy))
}

func profileWorkStateText(result coreeconomy.ProfileResult) string {
	if result.WorkUserFound && result.WorkUser.EndTimeUnix-result.NowUnix > 0 {
		return "打工中"
	}
	return "待業中"
}

func profileConfigIntText(value int64, found bool) string {
	if !found {
		return "無資料"
	}
	return coreeconomy.LegacyProfileAmount(float64(value))
}

func profileConfigFloatText(value float64, found bool) string {
	if !found {
		return "無資料"
	}
	return coreeconomy.LegacyProfileAmount(value)
}

func legacyProfileDate(value time.Time) string {
	if value.IsZero() {
		return "1970/01/01"
	}
	return value.Format("2006/01/02")
}

func drawProfileTextCentered(img *image.RGBA, x, y int, text string, c color.RGBA, scale int) {
	offset := estimatedProfileTextWidth(text, scale) / 2
	if face := signFontFace(float64(scale) * 10); face != nil {
		offset = font.MeasureString(face, text).Round() / 2
		_ = face.Close()
	}
	drawText(img, x-offset, y, text, c, scale)
}

func estimatedProfileTextWidth(text string, scale int) int {
	return len([]rune(text)) * (6*scale + scale)
}
