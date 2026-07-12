package retry

import (
	"errors"
	"time"
)

var ErrInvalidRetryAfter = errors.New("invalid Retry-After")

// ParseRetryAfter parses Retry-After into a duration bounded by max.
func ParseRetryAfter(value string, now time.Time, max time.Duration) (time.Duration, error) {
	panic("TODO")
}
