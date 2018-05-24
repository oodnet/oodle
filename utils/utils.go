package utils

import (
	"time"

	"github.com/hako/durafmt"
)

func FmtTime(t time.Time) string {
	// gets rid of milliseconds, I think?
	since := time.Since(t).Truncate(time.Second)
	// formats it to 1 day etc.
	return durafmt.Parse(since).String()
}

func FmtDur(d time.Duration) string {
	// gets rid of milliseconds, I think?
	d = d.Truncate(time.Second)
	// formats it to 1 second etc.
	return durafmt.Parse(d).String()
}
