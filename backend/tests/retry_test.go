package tests

import (
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
)

func TestPinRetryBackoff(t *testing.T) {
	for attempts, want := range map[int]time.Duration{1: time.Minute, 2: 2 * time.Minute, 4: 8 * time.Minute, 9: 256 * time.Minute, 10: attestation.RetryLongest} {
		if got := attestation.Backoff(attempts); got != want {
			t.Errorf("Backoff(%d) = %s, want %s", attempts, got, want)
		}
	}
}
