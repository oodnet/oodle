package plugins

import (
	"github.com/godwhoa/oodle/oodle"
)

type Echo struct {
	nick string
	oodle.BaseTrigger
}

func RegisterEcho(config *oodle.Config, bot oodle.Bot) {
	echo := &Echo{nick: config.Nick}
	bot.RegisterCommand(echo)
}

func (echo *Echo) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      "",
		Name:        echo.nick + "!",
		Description: "Exlamates your nick back!",
		Usage:       echo.nick + "!",
	}
}

func (echo *Echo) Execute(nick string, args []string) (string, error) {
	return nick + "!", nil
}
