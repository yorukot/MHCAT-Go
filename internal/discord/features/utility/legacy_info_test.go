package utility

import (
	"testing"
	"time"
)

func TestLegacyInfoJavaScriptFormatting(t *testing.T) {
	if got := formatLegacyShardUptime(27*time.Hour + 4*time.Minute + 5*time.Second + 999*time.Millisecond); got != "27h04m05s" {
		t.Fatalf("shard uptime = %q", got)
	}
	if got := formatLegacyShardUptime(-time.Second); got != "00h00m00s" {
		t.Fatalf("negative shard uptime = %q", got)
	}
	if got := formatLegacyShardPing(12*time.Millisecond + 999*time.Microsecond); got != "12" {
		t.Fatalf("shard ping = %q", got)
	}
	if got := formatLegacyShardPing(-time.Millisecond); got != "-1" {
		t.Fatalf("negative shard ping = %q", got)
	}

	beforeHalf := time.Unix(1_700_000_000, (499 * time.Millisecond).Nanoseconds())
	afterHalf := time.Unix(1_700_000_000, (500 * time.Millisecond).Nanoseconds())
	if got := legacyRoundedUnix(beforeHalf); got != 1_700_000_000 {
		t.Fatalf("rounded unix before half = %d", got)
	}
	if got := legacyRoundedUnix(afterHalf); got != 1_700_000_001 {
		t.Fatalf("rounded unix at half = %d", got)
	}
	if got := legacyBootUnix(afterHalf, 60*time.Second); got != 1_699_999_941 {
		t.Fatalf("rounded boot unix = %d", got)
	}

	if got := legacyLocale("zh-TW"); got != "🇹🇼`(zh-TW)`" {
		t.Fatalf("known locale = %q", got)
	}
	if got := legacyLocale("en-CA"); got != "undefined`(en-CA)`" {
		t.Fatalf("unknown locale = %q", got)
	}
}
