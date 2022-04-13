package utils

import "time"

type TimeoutWithStart struct {
	start    time.Time
	duration time.Duration
}

func (t *TimeoutWithStart) Elapsed() bool {
	return time.Since(t.start) > t.duration
}

func (t *TimeoutWithStart) Remaining() time.Duration {
	ret := t.duration - time.Since(t.start)
	if ret < 0 {
		ret = 0
	}
	return ret
}

func NewTimeoutWithStart(timeout time.Duration) *TimeoutWithStart {
	return &TimeoutWithStart{
		start:    time.Now().UTC(),
		duration: timeout,
	}
}
