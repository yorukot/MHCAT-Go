package domain

import "errors"

var ErrInvalidWorkPayout = errors.New("invalid work payout")
var ErrWorkPayoutStateConflict = errors.New("work payout state conflict")
var ErrWorkPayoutCoinConflict = errors.New("work payout coin conflict")
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
