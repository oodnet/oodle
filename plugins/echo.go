package plugins

import (
	"github.com/godwhoa/oodle/oodle"
)

type Echo struct {
	Nick string
	oodle.BaseTrigger
}

func (echo *Echo) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      "",
		Name:        echo.Nick + "!",
		Description: "Exlamates your nick back!",
		Usage:       echo.Nick + "!",
	}
}

func (echo *Echo) Execute(nick string, args []string) (string, error) {
	return nick + "!", nil
}
