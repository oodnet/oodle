package plugins

import (
	"time"

	"github.com/godwhoa/oodle/oodle"
)

type Seen struct {
	store map[string]time.Time
	oodle.BaseTrigger
}

func NewSeen(nick string) *Seen {
	return &Seen{}
}

func (seen *Seen) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      "!",
		Name:        "seen",
		Description: "Tells you when it last saw someone",
		Usage:       "%s <nick>",
	}
}

func (seen *Seen) Execute(nick string, args []string) (string, error) {
	return "", nil
}
