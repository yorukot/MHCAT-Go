package domain

import (
	"errors"
	"math"
	"time"
)

var ErrInvalidWorkPayout = errors.New("invalid work payout")
var ErrWorkPayoutStateConflict = errors.New("work payout state conflict")
var ErrWorkPayoutMarkerConflict = errors.New("work payout marker conflict")

type WorkPayoutResult struct {
	EligibleJobs       int64
	ProcessedJobs      int64
	CoinMatched        int64
	CoinModified       int64
	CoinUpserted       int64
	IdempotentReplays  int64
	StateMatched       int64
	StateModified      int64
	SkippedInvalidJobs int64
}

func LegacyRoundedWorkPayoutUnix(now time.Time) int64 {
	return int64(math.Round(float64(now.UnixNano()) / float64(time.Second)))
}
