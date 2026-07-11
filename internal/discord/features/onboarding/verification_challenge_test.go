package onboarding

import (
	"bytes"
	"context"
	"errors"
	"image/jpeg"
	mathrand "math/rand"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestVerificationCaptchaGeneratorPreservesLegacyShape(t *testing.T) {
	generated, err := (verificationCaptchaGenerator{}).Generate(context.Background())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(generated.Answer) != 6 {
		t.Fatalf("answer = %q, want six letters", generated.Answer)
	}
	for _, letter := range generated.Answer {
		if !strings.ContainsRune(verificationCaptchaAlphabet, letter) {
			t.Fatalf("answer %q contains unsupported letter %q", generated.Answer, letter)
		}
	}
	if generated.ImageName != "captcha.jpeg" {
		t.Fatalf("image name = %q", generated.ImageName)
	}
	image, err := jpeg.Decode(bytes.NewReader(generated.ImageData))
	if err != nil {
		t.Fatalf("decode JPEG: %v", err)
	}
	if bounds := image.Bounds(); bounds.Dx() != 400 || bounds.Dy() != 250 {
		t.Fatalf("image bounds = %v", bounds)
	}
	dark, nonWhite := 0, 0
	for y := image.Bounds().Min.Y; y < image.Bounds().Max.Y; y++ {
		for x := image.Bounds().Min.X; x < image.Bounds().Max.X; x++ {
			red, green, blue, _ := image.At(x, y).RGBA()
			if red+green+blue < 3*0x4000 {
				dark++
			}
			if red < 0xf000 || green < 0xf000 || blue < 0xf000 {
				nonWhite++
			}
		}
	}
	if dark < 1000 || nonWhite < 10000 {
		t.Fatalf("captcha lacks legacy noise: dark=%d nonWhite=%d", dark, nonWhite)
	}
}

func TestRandomVerificationAnswerUsesLegacyAlphabet(t *testing.T) {
	random := mathrand.New(mathrand.NewSource(42))
	for i := 0; i < 100; i++ {
		answer := randomVerificationAnswer(random)
		if len(answer) != 6 {
			t.Fatalf("answer = %q", answer)
		}
		for _, letter := range answer {
			if !strings.ContainsRune(verificationCaptchaAlphabet, letter) {
				t.Fatalf("answer %q contains unsupported letter %q", answer, letter)
			}
		}
	}
}

func TestVerificationCaptchaGeneratorHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (verificationCaptchaGenerator{}).Generate(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
}

func TestVerificationChallengeStoreExpiresAtTTLBoundary(t *testing.T) {
	now := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	store := newVerificationChallengeStore()
	store.now = func() time.Time { return now }
	challenge, err := store.Create(context.Background(), domain.VerificationChallenge{
		StateID: "state",
		GuildID: "guild",
		UserID:  "user",
		Answer:  "ABCDEF",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	now = now.Add(verificationChallengeTTL - time.Nanosecond)
	if got, err := store.Get(context.Background(), challenge.StateID); err != nil || got.Answer != "ABCDEF" {
		t.Fatalf("challenge before expiry = %#v, %v", got, err)
	}
	now = now.Add(time.Nanosecond)
	if _, err := store.Get(context.Background(), challenge.StateID); !errors.Is(err, domain.ErrInvalidVerificationChallenge) {
		t.Fatalf("expiry error = %v", err)
	}
}

func TestVerificationChallengeStoreClaimsOneCompletionAtATime(t *testing.T) {
	store := newVerificationChallengeStore()
	challenge, err := store.Create(context.Background(), domain.VerificationChallenge{
		StateID: "state",
		GuildID: "guild",
		UserID:  "user",
		Answer:  "ABCDEF",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := store.Claim(context.Background(), challenge.StateID); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if _, err := store.Claim(context.Background(), challenge.StateID); !errors.Is(err, domain.ErrInvalidVerificationChallenge) {
		t.Fatalf("concurrent claim error = %v", err)
	}
	store.Release(challenge.StateID)
	if _, err := store.Claim(context.Background(), challenge.StateID); err != nil {
		t.Fatalf("claim after release: %v", err)
	}
	store.Delete(challenge.StateID)
	if _, err := store.Get(context.Background(), challenge.StateID); !errors.Is(err, domain.ErrInvalidVerificationChallenge) {
		t.Fatalf("deleted challenge error = %v", err)
	}
}
