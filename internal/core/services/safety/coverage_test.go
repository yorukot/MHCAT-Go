package safety

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
)

func TestReportErrorClassifiers(t *testing.T) {
	if !IsInvalidReport(domain.ErrInvalidScamURLReport) || IsInvalidReport(errors.New("other")) {
		t.Fatal("invalid report classifier mismatch")
	}
	if !IsAlreadyReported(domain.ErrScamURLAlreadyReported) || IsAlreadyReported(errors.New("other")) {
		t.Fatal("already reported classifier mismatch")
	}
}
