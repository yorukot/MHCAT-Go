package economy

import (
	"context"
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestEconomyModuleMetadataAndLifecycleHelpers(t *testing.T) {
	module := NewModule(fakemongo.NewEconomyRepository(), nil, nil)
	if module.Name() != "economy-query" || len(module.Commands()) == 0 {
		t.Fatalf("economy metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
	module.RegisterEventRoutes(nil)
	if err := (Module{}).StopCoinGameLifecycle(context.Background()); err != nil {
		t.Fatalf("stop empty lifecycle: %v", err)
	}
	if value := legacyRandomInt(0); value != 0 {
		t.Fatalf("zero random int = %d", value)
	}
	if value := legacyRandomInt(4); value < 0 || value >= 4 {
		t.Fatalf("random int = %d", value)
	}
}

func TestCoinGameTimeoutCancel(t *testing.T) {
	manager := newCoinGameTimeoutManager(nil)
	manager.Schedule("game", 1, time.Now().Add(time.Hour), func(context.Context) {})
	manager.Cancel("game")
	manager.Cancel("")
	if err := manager.Stop(context.Background()); err != nil {
		t.Fatalf("stop timeout manager: %v", err)
	}
}

func TestEconomyRenderingFallbackHelpers(t *testing.T) {
	face := &coinRankFallbackFace{}
	if _, _, ok := face.GlyphBounds('x'); ok {
		t.Fatal("empty fallback face reported glyph bounds")
	}
	if metrics := face.Metrics(); metrics.Height != 0 {
		t.Fatalf("empty fallback metrics = %#v", metrics)
	}
	if got := truncateRunes("abcdef", 3); got != "abc" {
		t.Fatalf("truncated runes = %q", got)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 50, 50))
	source := image.NewRGBA(image.Rect(0, 0, 10, 10))
	source.Set(5, 5, color.RGBA{R: 255, A: 255})
	drawRoundedImage(canvas, source, image.Pt(2, 2), 2)
	if _, _, _, alpha := canvas.At(7, 7).RGBA(); alpha == 0 {
		t.Fatal("rounded image did not copy center pixel")
	}
	drawCheck(canvas, 0, 0)
	if _, _, _, alpha := canvas.At(1, 21).RGBA(); alpha == 0 {
		t.Fatal("check mark was not drawn")
	}
}

func TestEconomyMessageAndTextHelpers(t *testing.T) {
	if profileConfigIntText(10, true) == "" || profileConfigIntText(0, false) != "無資料" {
		t.Fatal("profile integer text mismatch")
	}
	if profileConfigFloatText(1.5, true) == "" || profileConfigFloatText(0, false) != "無資料" {
		t.Fatal("profile float text mismatch")
	}
	if message := shopDeleteSuccessMessage(); len(message.Embeds) != 1 {
		t.Fatalf("shop delete message = %#v", message)
	}
	if message := shopDetailError(ports.ErrShopItemMissing); len(message.Embeds) != 1 || len(message.Components) != 0 {
		t.Fatalf("shop detail error = %#v", message)
	}
	if got := legacySignCustomID("user", 2026, time.July); got != "/user_sing{2026}-[7]" {
		t.Fatalf("legacy sign id = %q", got)
	}
}
