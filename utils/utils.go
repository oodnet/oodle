package utils

import (
	"strings"
	"time"

	"github.com/godwhoa/oodle/oodle"
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

// Alias returns a new command with replaced prefix, name and usage
func Alias(prefix, name string, cmd oodle.Command) oodle.Command {
	firstFragment := strings.Split(cmd.Usage, " ")[0] // usually it's command's prefix+name
	usage := strings.Replace(cmd.Usage, firstFragment, prefix+name, 1)
	cmd.Prefix = prefix
	cmd.Name = name
	cmd.Usage = usage
	return cmd
}
