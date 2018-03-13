package plugins

import (
	"strings"

	"github.com/godwhoa/oodle/oodle"
)

type Echo struct {
	nick      string
	customCmd map[string]string
	oodle.BaseInteractive
}

func (e *Echo) Configure(config *oodle.Config) {
	e.nick = config.Nick
	e.customCmd = config.CustomCommands
}

func (e *Echo) OnEvent(event interface{}) {
	message, ok := event.(oodle.Message)
	if !ok {
		return
	}
	args := strings.Split(strings.TrimSpace(message.Msg), " ")
	if len(args) < 1 {
		return
	}
	if msg, ok := e.customCmd[args[0]]; ok {
		e.IRC.Send(msg)
	}
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
