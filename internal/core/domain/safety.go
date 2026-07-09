package domain

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidAntiScamConfig = errors.New("invalid anti-scam config")
var ErrInvalidScamURLReport = errors.New("invalid scam URL report")
var ErrScamURLAlreadyReported = errors.New("scam URL already reported")

type AntiScamConfig struct {
	GuildID string
	Open    bool
}

func (c AntiScamConfig) Validate() error {
	if strings.TrimSpace(c.GuildID) == "" {
		return ErrInvalidAntiScamConfig
	}
	return nil
}

type ScamURLReport struct {
	URL            string
	ReporterUserID string
}

func (r ScamURLReport) Validate() error {
	if !LooksLikeURL(r.URL) {
		return ErrInvalidScamURLReport
	}
	if strings.TrimSpace(r.ReporterUserID) == "" {
		return ErrInvalidScamURLReport
	}
	return nil
}

func LooksLikeURL(raw string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || parsed == nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
