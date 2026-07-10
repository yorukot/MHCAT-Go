package domain

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf16"
)

var ErrInvalidAntiScamConfig = errors.New("invalid anti-scam config")
var ErrInvalidScamURLReport = errors.New("invalid scam URL report")
var ErrScamURLAlreadyReported = errors.New("scam URL already reported")

var (
	legacyURLProtocolAndDomain = regexp.MustCompile(`^(?:[A-Za-z0-9_]+:)?//(\S+)$`)
	legacyLocalhostDomain      = regexp.MustCompile(`^localhost[:?0-9]*(?:[^:?0-9].*)?$`)
)

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
	match := legacyURLProtocolAndDomain.FindStringSubmatch(raw)
	if len(match) != 2 || match[1] == "" {
		return false
	}
	domain := match[1]
	if strings.IndexFunc(domain, legacyURLWhitespace) >= 0 {
		return false
	}
	if legacyLocalhostDomain.MatchString(domain) {
		return true
	}
	dot := strings.IndexByte(domain, '.')
	return dot > 0 && len(utf16.Encode([]rune(domain[dot+1:]))) >= 2
}

func legacyURLWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', '\u00a0', '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000', '\ufeff':
		return true
	default:
		return character >= '\u2000' && character <= '\u200a'
	}
}
