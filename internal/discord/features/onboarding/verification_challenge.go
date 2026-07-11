package onboarding

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/math/fixed"
)

const verificationChallengeTTL = 5 * time.Minute

type verificationChallengeStore struct {
	mu     sync.Mutex
	now    func() time.Time
	ttl    time.Duration
	values map[string]storedVerificationChallenge
}

type storedVerificationChallenge struct {
	challenge domain.VerificationChallenge
	expiresAt time.Time
	claimed   bool
}

func newVerificationChallengeStore() *verificationChallengeStore {
	return &verificationChallengeStore{
		now:    time.Now,
		ttl:    verificationChallengeTTL,
		values: map[string]storedVerificationChallenge{},
	}
}

func (s *verificationChallengeStore) Create(ctx context.Context, challenge domain.VerificationChallenge) (domain.VerificationChallenge, error) {
	if err := ctx.Err(); err != nil {
		return domain.VerificationChallenge{}, err
	}
	if s == nil {
		return domain.VerificationChallenge{}, domain.ErrInvalidVerificationChallenge
	}
	challenge.GuildID = strings.TrimSpace(challenge.GuildID)
	challenge.UserID = strings.TrimSpace(challenge.UserID)
	challenge.Answer = strings.TrimSpace(challenge.Answer)
	if challenge.StateID == "" {
		stateID, err := randomStateID()
		if err != nil {
			return domain.VerificationChallenge{}, err
		}
		challenge.StateID = stateID
	}
	if err := challenge.Validate(); err != nil {
		return domain.VerificationChallenge{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	s.values[challenge.StateID] = storedVerificationChallenge{challenge: challenge, expiresAt: s.now().Add(s.ttl)}
	return challenge, nil
}

func (s *verificationChallengeStore) Get(ctx context.Context, stateID string) (domain.VerificationChallenge, error) {
	if err := ctx.Err(); err != nil {
		return domain.VerificationChallenge{}, err
	}
	if s == nil {
		return domain.VerificationChallenge{}, domain.ErrInvalidVerificationChallenge
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	stored, ok := s.values[strings.TrimSpace(stateID)]
	if !ok {
		return domain.VerificationChallenge{}, domain.ErrInvalidVerificationChallenge
	}
	return stored.challenge, nil
}

func (s *verificationChallengeStore) Claim(ctx context.Context, stateID string) (domain.VerificationChallenge, error) {
	if err := ctx.Err(); err != nil {
		return domain.VerificationChallenge{}, err
	}
	if s == nil {
		return domain.VerificationChallenge{}, domain.ErrInvalidVerificationChallenge
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	key := strings.TrimSpace(stateID)
	stored, ok := s.values[key]
	if !ok || stored.claimed {
		return domain.VerificationChallenge{}, domain.ErrInvalidVerificationChallenge
	}
	stored.claimed = true
	s.values[key] = stored
	return stored.challenge, nil
}

func (s *verificationChallengeStore) Release(stateID string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.TrimSpace(stateID)
	stored, ok := s.values[key]
	if !ok {
		return
	}
	stored.claimed = false
	s.values[key] = stored
}

func (s *verificationChallengeStore) Delete(stateID string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, strings.TrimSpace(stateID))
}

func (s *verificationChallengeStore) pruneLocked() {
	now := s.now()
	for id, stored := range s.values {
		if !stored.expiresAt.IsZero() && !now.Before(stored.expiresAt) {
			delete(s.values, id)
		}
	}
}

func randomStateID() (string, error) {
	var raw [10]byte
	if _, err := cryptorand.Read(raw[:]); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw[:]), "="), nil
}

type verificationCaptchaGenerator struct{}

func (verificationCaptchaGenerator) Generate(ctx context.Context) (coreservice.VerificationGeneratedChallenge, error) {
	if err := ctx.Err(); err != nil {
		return coreservice.VerificationGeneratedChallenge{}, err
	}
	random, err := newVerificationCaptchaRandom()
	if err != nil {
		return coreservice.VerificationGeneratedChallenge{}, err
	}
	answer := randomVerificationAnswer(random)
	imageData, err := renderCaptchaJPEG(answer, random)
	if err != nil {
		return coreservice.VerificationGeneratedChallenge{}, err
	}
	return coreservice.VerificationGeneratedChallenge{Answer: answer, ImageName: "captcha.jpeg", ImageData: imageData}, nil
}

func newVerificationCaptchaRandom() (*mathrand.Rand, error) {
	var seed [8]byte
	if _, err := cryptorand.Read(seed[:]); err != nil {
		return nil, err
	}
	return mathrand.New(mathrand.NewSource(int64(binary.LittleEndian.Uint64(seed[:])))), nil
}

const verificationCaptchaAlphabet = "ABCDEFHIJLMNOPSTUVWXYZ"

func randomVerificationAnswer(random *mathrand.Rand) string {
	var answer [6]byte
	for index := range answer {
		answer[index] = verificationCaptchaAlphabet[random.Intn(len(verificationCaptchaAlphabet))]
	}
	return string(answer[:])
}

func renderCaptchaJPEG(answer string, random *mathrand.Rand) ([]byte, error) {
	const width = 400
	const height = 250
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	coords := [4][5]int{}
	for i := range coords {
		for j := range coords[i] {
			coords[i][j] = int(math.Round(random.Float64()*80)) + j*80
		}
		if i%2 == 0 {
			random.Shuffle(len(coords[i]), func(left int, right int) {
				coords[i][left], coords[i][right] = coords[i][right], coords[i][left]
			})
		}
	}
	black := color.RGBA{A: 255}
	for i := 0; i < 5; i++ {
		drawCaptchaLine(canvas, coords[0][i], 0, coords[1][i], 400, 4, black)
		drawCaptchaLine(canvas, 0, coords[2][i], 400, coords[3][i], 4, black)
	}
	for i := 0; i < 200; i++ {
		x := int(math.Round(random.Float64()*360)) + 20
		y := int(math.Round(random.Float64()*360)) + 20
		radius := int(math.Round(random.Float64()*7)) + 1
		drawCaptchaCircle(canvas, x, y, float64(radius), black)
	}

	drawCaptchaAnswer(canvas, answer, random)
	for i := 0; i < 5000; i++ {
		pointColor := color.RGBA{
			R: uint8(random.Intn(256)),
			G: uint8(random.Intn(256)),
			B: uint8(random.Intn(256)),
			A: 160,
		}
		drawCaptchaCircle(canvas, random.Intn(width+1), random.Intn(height+1), random.Float64()*2, pointColor)
	}

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, canvas, &jpeg.Options{Quality: 75}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func drawCaptchaAnswer(canvas *image.RGBA, answer string, random *mathrand.Rand) {
	layer := image.NewRGBA(image.Rect(0, 0, 400, 130))
	face := verificationCaptchaFontFace(90)
	if face != nil {
		defer face.Close()
		width := font.MeasureString(face, answer).Ceil()
		baseline := face.Metrics().Ascent.Ceil()
		drawer := &font.Drawer{
			Dst:  layer,
			Src:  image.NewUniform(color.Black),
			Face: face,
			Dot:  fixed.P((layer.Bounds().Dx()-width)/2, baseline),
		}
		drawer.DrawString(answer)
		drawer.Dot.X += fixed.I(1)
		drawer.DrawString(answer)
	} else {
		small := image.NewRGBA(image.Rect(0, 0, 48, 15))
		drawer := &font.Drawer{Dst: small, Src: image.NewUniform(color.Black), Face: basicfont.Face7x13, Dot: fixed.P(3, 13)}
		drawer.DrawString(answer)
		xdraw.NearestNeighbor.Scale(layer, image.Rect(45, 10, 355, 105), small, small.Bounds(), draw.Over, nil)
	}

	centerX := int(math.Round(random.Float64()*100-50)) + 200
	topY := 125 - int(math.Round(random.Float64()*62.5-31.25))
	angle := random.Float64() - 0.5
	cosine, sine := math.Cos(angle), math.Sin(angle)
	sourcePivotX := 200.0
	matrix := f64.Aff3{
		cosine, -sine, float64(centerX) - cosine*sourcePivotX,
		sine, cosine, float64(topY) - sine*sourcePivotX,
	}
	xdraw.CatmullRom.Transform(canvas, matrix, layer, layer.Bounds(), draw.Over, nil)
}

func drawCaptchaLine(canvas *image.RGBA, x0 int, y0 int, x1 int, y1 int, width int, lineColor color.RGBA) {
	steps := max(absInt(x1-x0), absInt(y1-y0))
	if steps == 0 {
		drawCaptchaCircle(canvas, x0, y0, float64(width)/2, lineColor)
		return
	}
	for step := 0; step <= steps; step++ {
		x := x0 + (x1-x0)*step/steps
		y := y0 + (y1-y0)*step/steps
		drawCaptchaCircle(canvas, x, y, float64(width)/2, lineColor)
	}
}

func drawCaptchaCircle(canvas *image.RGBA, centerX int, centerY int, radius float64, circleColor color.RGBA) {
	limit := int(math.Ceil(radius))
	radiusSquared := radius * radius
	for y := centerY - limit; y <= centerY+limit; y++ {
		for x := centerX - limit; x <= centerX+limit; x++ {
			dx, dy := float64(x-centerX), float64(y-centerY)
			if dx*dx+dy*dy <= radiusSquared {
				blendCaptchaPixel(canvas, x, y, circleColor)
			}
		}
	}
}

func blendCaptchaPixel(canvas *image.RGBA, x int, y int, source color.RGBA) {
	if !image.Pt(x, y).In(canvas.Bounds()) {
		return
	}
	if source.A == 255 {
		canvas.SetRGBA(x, y, source)
		return
	}
	destination := canvas.RGBAAt(x, y)
	alpha := uint32(source.A)
	inverse := uint32(255 - source.A)
	canvas.SetRGBA(x, y, color.RGBA{
		R: uint8((uint32(source.R)*alpha + uint32(destination.R)*inverse) / 255),
		G: uint8((uint32(source.G)*alpha + uint32(destination.G)*inverse) / 255),
		B: uint8((uint32(source.B)*alpha + uint32(destination.B)*inverse) / 255),
		A: 255,
	})
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

var verificationCaptchaFont struct {
	once sync.Once
	font *opentype.Font
}

func verificationCaptchaFontFace(size float64) font.Face {
	verificationCaptchaFont.once.Do(func() {
		for _, relative := range []string{
			"node_modules/@haileybot/captcha-generator/assets/Swift.ttf",
			"fonts/Comic-Sans-MS-copy-5-.ttf",
			"fonts/TaipeiSansTCBeta-Regular.ttf",
		} {
			for _, path := range verificationAssetCandidates(relative) {
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				parsed, err := opentype.Parse(data)
				if err == nil {
					verificationCaptchaFont.font = parsed
					return
				}
			}
		}
	})
	if verificationCaptchaFont.font == nil {
		return nil
	}
	face, err := opentype.NewFace(verificationCaptchaFont.font, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}
	return face
}

func verificationAssetCandidates(relative string) []string {
	return []string{
		relative,
		filepath.Join("MHCAT", relative),
		filepath.Join("..", "MHCAT", relative),
		filepath.Join("..", "..", "..", "..", "..", "MHCAT", relative),
	}
}

var _ coreservice.VerificationChallengeStore = (*verificationChallengeStore)(nil)
var _ coreservice.VerificationChallengeGenerator = verificationCaptchaGenerator{}
