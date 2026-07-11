package lottery

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestLotteryCombinedModuleMetadataAndRandomIndex(t *testing.T) {
	module := NewModuleWithComponents(fakemongo.NewLotteryRepository(), nil, nil, nil, nil, nil)
	if module.Name() != "lottery-disabled-command" || len(module.Commands()) == 0 {
		t.Fatalf("lottery metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
	if index, err := lotteryCryptoRandomIndex(0); err != nil || index != 0 {
		t.Fatalf("zero random index=%d err=%v", index, err)
	}
	if index, err := lotteryCryptoRandomIndex(4); err != nil || index < 0 || index >= 4 {
		t.Fatalf("random index=%d err=%v", index, err)
	}
}

func TestLotteryComponentErrorHelpers(t *testing.T) {
	lookup := lotteryLookupErrorMessage(ports.ErrLotteryNotFound, true)
	generic := lotteryComponentErrorMessage("failure")
	if !lookup.Ephemeral || lookup.Content == "" || !generic.Ephemeral || generic.Content == "" {
		t.Fatalf("lookup=%#v generic=%#v", lookup, generic)
	}
	_ = lotteryLookupErrorMessage(errors.New("unknown"), false)
}
