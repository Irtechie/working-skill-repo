package retry

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidRetryAfter = errors.New("invalid Retry-After")

func ParseRetryAfter(value string, now time.Time, max time.Duration) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" || max < 0 {
		return 0, ErrInvalidRetryAfter
	}
	if seconds, err := strconv.ParseUint(value, 10, 64); err == nil {
		if seconds > uint64(math.MaxInt64/int64(time.Second)) {
			return 0, ErrInvalidRetryAfter
		}
		duration := time.Duration(seconds) * time.Second
		if duration > max {
			return max, nil
		}
		return duration, nil
	}
	date, err := http.ParseTime(value)
	if err != nil {
		date, err = time.Parse(time.RFC1123, value)
		if err != nil {
			return 0, ErrInvalidRetryAfter
		}
	}
	if !date.After(now) {
		return 0, nil
	}
	duration := date.Sub(now)
	if duration > max {
		return max, nil
	}
	return duration, nil
}
