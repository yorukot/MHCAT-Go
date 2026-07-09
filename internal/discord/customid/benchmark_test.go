package customid_test

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/customid"
)

func BenchmarkParseVersionedID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = customid.ParseComponent("mhcat:v1:poll:vote:opt_1")
	}
}

func BenchmarkParseLegacyID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = customid.ParseComponent("[123456789012345678]{2}text_rank")
	}
}
