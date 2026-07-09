package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type DiscordWebhookReporter struct {
	Client *http.Client
	URL    string
}

func NewDiscordWebhookReporter(url string) DiscordWebhookReporter {
	return DiscordWebhookReporter{
		Client: &http.Client{Timeout: 10 * time.Second},
		URL:    strings.TrimSpace(url),
	}
}

func (r DiscordWebhookReporter) SendScamURLReport(ctx context.Context, report domain.ScamURLReport) error {
	if err := report.Validate(); err != nil {
		return err
	}
	endpoint := strings.TrimSpace(r.URL)
	if endpoint == "" {
		return domain.ErrInvalidScamURLReport
	}
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	payload := map[string]string{
		"content": fmt.Sprintf("```%s```\nby:<@%s>", sanitizeDiscordCodeBlock(report.URL), report.ReporterUserID),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode discord webhook report: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create discord webhook report: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("send discord webhook report: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("discord webhook report status %d", response.StatusCode)
	}
	return ctx.Err()
}

func sanitizeDiscordCodeBlock(value string) string {
	return strings.ReplaceAll(value, "```", "`\u200b``")
}

var _ ports.ScamReportSender = DiscordWebhookReporter{}
