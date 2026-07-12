package retry

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		value string
		max   time.Duration
		want  time.Duration
		err   bool
	}{
		{"seconds", "120", 10 * time.Minute, 2 * time.Minute, false},
		{"zero", "0", 10 * time.Minute, 0, false},
		{"trim", " 15 ", time.Minute, 15 * time.Second, false},
		{"clamp seconds", "999", 30 * time.Second, 30 * time.Second, false},
		{"http date", now.Add(45 * time.Second).Format(time.RFC1123), time.Minute, 45 * time.Second, false},
		{"past date", now.Add(-time.Minute).Format(time.RFC1123), time.Hour, 0, false},
		{"empty", "", time.Minute, 0, true},
		{"negative", "-1", time.Minute, 0, true},
		{"signed", "+1", time.Minute, 0, true},
		{"fractional", "1.5", time.Minute, 0, true},
		{"overflow", "18446744073709551616", time.Hour, 0, true},
		{"malformed", "tomorrow", time.Hour, 0, true},
		{"bad max", "1", -time.Second, 0, true},
		{"duration overflow", "9223372037", time.Duration(math.MaxInt64), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRetryAfter(tt.value, now, tt.max)
			if (err != nil) != tt.err {
				t.Fatalf("error=%v, want error=%v", err, tt.err)
			}
			if tt.err && !errors.Is(err, ErrInvalidRetryAfter) {
				t.Fatalf("error=%v, want ErrInvalidRetryAfter", err)
			}
			if got != tt.want {
				t.Fatalf("duration=%v, want %v", got, tt.want)
			}
		})
	}
}
