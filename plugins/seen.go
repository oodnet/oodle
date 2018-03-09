package plugins

import (
	"fmt"
	"time"

	"github.com/godwhoa/oodle/oodle"
)

type Seen struct {
	store map[string]time.Time
	oodle.BaseTrigger
}

func (seen *Seen) Configure(config *oodle.Config) {
	seen.store = make(map[string]time.Time)
}

func (seen *Seen) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "seen",
		Description: "Tells you when it last saw someone",
		Usage:       ".seen <nick>",
	}
}

func (seen *Seen) OnEvent(event interface{}) {
	switch event.(type) {
	case oodle.Join:
		seen.store[event.(oodle.Join).Nick] = time.Now()
	case oodle.Leave:
		seen.store[event.(oodle.Leave).Nick] = time.Now()
	case oodle.Message:
		seen.store[event.(oodle.Message).Nick] = time.Now()
	}
}

func (seen *Seen) Execute(nick string, args []string) (string, error) {
	if len(args) < 1 {
		return "", oodle.ErrUsage
	}
	if lastSeen, ok := seen.store[args[0]]; ok {
		return fmt.Sprintf("%s was last seen %s ago.", args[0], fmtTime(lastSeen)), nil
	}
	return fmt.Sprintf("No logs for %s", args[0]), nil
}
