package safety

import (
	"context"
	"errors"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type ReportService struct {
	catalog ports.ScamURLCatalog
	sender  ports.ScamReportSender
}

func NewReportService(catalog ports.ScamURLCatalog, sender ports.ScamReportSender) ReportService {
	return ReportService{catalog: catalog, sender: sender}
}

func (s ReportService) Report(ctx context.Context, rawURL string, reporterUserID string) (domain.ScamURLReport, error) {
	if err := ctx.Err(); err != nil {
		return domain.ScamURLReport{}, err
	}
	if s.catalog == nil || s.sender == nil {
		return domain.ScamURLReport{}, domain.ErrInvalidScamURLReport
	}
	report := domain.ScamURLReport{
		URL:            strings.TrimSpace(rawURL),
		ReporterUserID: strings.TrimSpace(reporterUserID),
	}
	if err := report.Validate(); err != nil {
		return domain.ScamURLReport{}, err
	}
	exists, err := s.catalog.ContainsScamURL(ctx, report.URL)
	if err != nil {
		return domain.ScamURLReport{}, err
	}
	if exists {
		return domain.ScamURLReport{}, domain.ErrScamURLAlreadyReported
	}
	if err := s.sender.SendScamURLReport(ctx, report); err != nil {
		return domain.ScamURLReport{}, err
	}
	if err := ctx.Err(); err != nil {
		return domain.ScamURLReport{}, err
	}
	return report, nil
}

func IsInvalidReport(err error) bool {
	return errors.Is(err, domain.ErrInvalidScamURLReport)
}

func IsAlreadyReported(err error) bool {
	return errors.Is(err, domain.ErrScamURLAlreadyReported)
}
