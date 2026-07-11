package gacha

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestGachaModuleMetadata(t *testing.T) {
	repository := fakemongo.NewGachaRepository()
	module := NewModuleWithRepositories(repository, repository, repository, repository, repository, nil, nil, nil, nil)
	if module.Name() != "gacha" || len(module.Commands()) == 0 {
		t.Fatalf("gacha metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}
