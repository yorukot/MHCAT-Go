package economy

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	coreeconomy "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/economy"
	xdraw "golang.org/x/image/draw"

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
	displayName := truncateLegacyCoinRankText(view.DisplayName)
	if displayName == "" {
		displayName = view.Result.UserID
	}
	drawCoinRankText(canvas, 151, 80, displayName, color.RGBA{R: 211, G: 211, B: 211, A: 255}, 45)
	drawCoinRankText(canvas, 151, 120, view.GuildName, color.RGBA{R: 168, G: 168, B: 168, A: 255}, 25)
	drawCoinRankNumericText(canvas, 1220, 100, legacyProfileDate(view.MemberJoinedAt), color.RGBA{R: 211, G: 211, B: 211, A: 255}, 40)
	drawCoinRankNumericText(canvas, 960, 100, legacyProfileDate(view.UserCreatedAt), color.RGBA{R: 211, G: 211, B: 211, A: 255}, 40)
}

func drawProfileStats(canvas *image.RGBA, result coreeconomy.ProfileResult) {
	drawProfileProgress(canvas, 550, 333, profileProgressWidth(result.TextXP, result.TextXPFound, false), color.RGBA{R: 100, G: 255, B: 191, A: 255})
	drawProfileProgress(canvas, 1038, 333, profileProgressWidth(result.VoiceXP, result.VoiceXPFound, true), color.RGBA{R: 234, G: 121, B: 255, A: 255})
	drawProfileXPTextCentered(canvas, 750, 363, profileXPProgressText(result.TextXP, result.TextXPFound, false), color.RGBA{R: 255, G: 88, B: 9, A: 255}, 30)
	drawProfileXPTextCentered(canvas, 1238, 363, profileXPProgressText(result.VoiceXP, result.VoiceXPFound, true), color.RGBA{R: 40, G: 255, B: 40, A: 255}, 30)

	drawProfileTextCentered(canvas, 367, 243, profileRankText(result.TextRank, result.TextXPFound, true), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 367, 306, profileRankText(result.VoiceRank, result.VoiceXPFound, false), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 367, 369, profileRankText(result.CoinRank, result.CoinFound, false), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)

	drawProfileTextCentered(canvas, 864, 243, profileXPValue(result.TextXP, result.TextXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 864, 306, profileLevelValue(result.TextXP, result.TextXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 1351, 243, profileXPValue(result.VoiceXP, result.VoiceXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 1351, 306, profileLevelValue(result.VoiceXP, result.VoiceXPFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)

	drawProfileTextCentered(canvas, 295, 525, profileCoinText(result), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 295, 587, result.SignStatus, color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 639, 525, profileWorkEnergyText(result), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 639, 587, profileWorkStateText(result), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 1045, 525, profileConfigRawText(result.Config.GachaCostText, float64(result.Config.GachaCost), result.ConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 1045, 587, profileConfigRawText(result.Config.SignCoinsText, float64(result.Config.SignCoins), result.ConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 1385, 525, profileConfigRawText(result.WorkConfig.DailyEnergyText, float64(result.WorkConfig.DailyEnergy), result.WorkConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 1385, 587, profileConfigRawText(result.WorkConfig.MaxEnergyText, float64(result.WorkConfig.MaxEnergy), result.WorkConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
	drawProfileTextCentered(canvas, 1237, 652, profileConfigRawText(result.Config.XPMultipleText, result.Config.XPMultiple, result.ConfigFound), color.RGBA{R: 252, G: 252, B: 252, A: 255}, 40)
}

func drawProfileProgress(canvas *image.RGBA, x, y, width int, c color.RGBA) {
	if width <= 0 {
		return
	}
	for py := 0; py < 35; py++ {
		for px := 0; px < width; px++ {
			if insideRoundedRect(px, py, width, 35, 17) {
				canvas.Set(x+px, y+py, c)
			}
		}
	}
}

func profileProgressWidth(profile domain.XPProfile, found bool, voice bool) int {
	if !found {
		return 0
	}
	xp := coreeconomy.LegacyProfileXPNumber(profile)
	required := coreeconomy.LegacyProfileXPRequiredForProfile(profile, voice)
	if math.IsNaN(xp) || math.IsNaN(required) || math.IsInf(xp, 0) || math.IsInf(required, 0) || required == 0 {
		return 0
	}
	percent := int(math.Floor(xp/required*100 + 0.5))
	if percent <= 0 {
		return 0
	}
	return 35 + percent*36/10
}

func profileXPProgressText(profile domain.XPProfile, found bool, voice bool) string {
	if !found {
		return "0/0"
	}
	return coreeconomy.LegacyProfileXPAmount(profile) + "/" + coreeconomy.LegacyProfileAmount(coreeconomy.LegacyProfileXPRequiredForProfile(profile, voice))
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
	return coreeconomy.LegacyProfileXPAmount(profile)
}

func profileLevelValue(profile domain.XPProfile, found bool) string {
	if !found {
		return "沒有資料"
	}
	return coreeconomy.LegacyProfileLevelText(profile)
}

func profileCoinText(result coreeconomy.ProfileResult) string {
	if !result.CoinFound {
		return "0"
	}
	return coreeconomy.LegacyProfileCoinAmount(result.CoinBalance)
}

func profileWorkEnergyText(result coreeconomy.ProfileResult) string {
	if !result.WorkUserFound {
		return "0"
	}
	return coreeconomy.LegacyProfileRawAmount(result.WorkUser.EnergyText, float64(result.WorkUser.Energy))
}

func profileWorkStateText(result coreeconomy.ProfileResult) string {
	if !result.WorkUserFound {
		return "待業中"
	}
	return coreeconomy.LegacyProfileWorkState(result.WorkUser.EndTimeText, result.WorkUser.EndTimeUnix, result.NowUnix)
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

func profileConfigRawText(text string, fallback float64, found bool) string {
	if !found {
		return "無資料"
	}
	return coreeconomy.LegacyProfileRawAmount(text, fallback)
}

func legacyProfileDate(value time.Time) string {
	if value.IsZero() {
		return "1970/01/01"
	}
	return value.Format("2006/01/02")
}

func drawProfileTextCentered(img *image.RGBA, x, y int, text string, c color.RGBA, size int) {
	drawCoinRankCenteredNumericText(img, x, y, text, c, size)
}
