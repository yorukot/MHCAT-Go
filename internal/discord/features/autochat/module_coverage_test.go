package autochat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAutoChatModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewAutoChatConfigRepository(), nil)
	if module.Name() != "autochat-config" || len(module.Commands()) == 0 {
		t.Fatalf("autochat metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}

func TestAutoChatReplyWaitHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForAutoChatReply(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("wait error = %v", err)
	}
}
