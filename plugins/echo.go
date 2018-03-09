package plugins

import (
	"github.com/godwhoa/oodle/oodle"
)

type Echo struct {
	nick string
}

func (e *Echo) Configure(config *oodle.Config) {
	e.nick = config.Nick
}

func (e *Echo) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      "",
		Name:        e.nick + "!",
		Description: "Exlamates your nick back!",
		Usage:       e.nick + "!",
	}
}

func (e *Echo) Execute(nick string, args []string) (string, error) {
	return nick + "!", nil
}
