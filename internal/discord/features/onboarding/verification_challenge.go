package onboarding

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	coreservice "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/onboarding"
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

func (s *verificationChallengeStore) Delete(ctx context.Context, stateID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return domain.ErrInvalidVerificationChallenge
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, strings.TrimSpace(stateID))
	return nil
}

func (s *verificationChallengeStore) pruneLocked() {
	now := s.now()
	for id, stored := range s.values {
		if !stored.expiresAt.IsZero() && now.After(stored.expiresAt) {
			delete(s.values, id)
		}
	}
}

func randomStateID() (string, error) {
	var raw [10]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw[:]), "="), nil
}

type verificationCaptchaGenerator struct{}

func (verificationCaptchaGenerator) Generate(ctx context.Context) (coreservice.VerificationGeneratedChallenge, error) {
	if err := ctx.Err(); err != nil {
		return coreservice.VerificationGeneratedChallenge{}, err
	}
	answer, err := randomDigits(4)
	if err != nil {
		return coreservice.VerificationGeneratedChallenge{}, err
	}
	imageData, err := renderCaptchaJPEG(answer)
	if err != nil {
		return coreservice.VerificationGeneratedChallenge{}, err
	}
	return coreservice.VerificationGeneratedChallenge{Answer: answer, ImageName: "captcha.jpeg", ImageData: imageData}, nil
}

func randomDigits(count int) (string, error) {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		value, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + value.Int64()))
	}
	return builder.String(), nil
}

func renderCaptchaJPEG(answer string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 180, 64))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 245, G: 248, B: 255, A: 255}}, image.Point{}, draw.Src)
	for index, digit := range answer {
		drawSevenSegment(img, 18+index*38, 10, digit, color.RGBA{R: 45, G: 80, B: 180, A: 255})
	}
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, img, &jpeg.Options{Quality: 88}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

var sevenSegmentDigits = map[rune][7]bool{
	'0': {true, true, true, true, true, true, false},
	'1': {false, true, true, false, false, false, false},
	'2': {true, true, false, true, true, false, true},
	'3': {true, true, true, true, false, false, true},
	'4': {false, true, true, false, false, true, true},
	'5': {true, false, true, true, false, true, true},
	'6': {true, false, true, true, true, true, true},
	'7': {true, true, true, false, false, false, false},
	'8': {true, true, true, true, true, true, true},
	'9': {true, true, true, true, false, true, true},
}

func drawSevenSegment(img *image.RGBA, x int, y int, digit rune, c color.Color) {
	segments, ok := sevenSegmentDigits[digit]
	if !ok {
		return
	}
	rects := []image.Rectangle{
		image.Rect(x+4, y, x+28, y+5),
		image.Rect(x+28, y+4, x+33, y+24),
		image.Rect(x+28, y+28, x+33, y+48),
		image.Rect(x+4, y+48, x+28, y+53),
		image.Rect(x, y+28, x+5, y+48),
		image.Rect(x, y+4, x+5, y+24),
		image.Rect(x+4, y+24, x+28, y+29),
	}
	for index, enabled := range segments {
		if enabled {
			draw.Draw(img, rects[index], &image.Uniform{C: c}, image.Point{}, draw.Src)
		}
	}
}

var _ coreservice.VerificationChallengeStore = (*verificationChallengeStore)(nil)
var _ coreservice.VerificationChallengeGenerator = verificationCaptchaGenerator{}
