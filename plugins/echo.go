package plugins

import (
	"fmt"

	"github.com/godwhoa/oodle/oodle"
)

type Echo struct {
	nick string
	oodle.BaseTrigger
}

func RegisterEcho(config *oodle.Config, bot oodle.Bot) {
	echo := &Echo{nick: config.Nick}
	bot.RegisterCommand(echo)
	bot.RegisterTrigger(echo)
}

func (echo *Echo) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      "",
		Name:        echo.nick + "!",
		Description: "Exlamates your nick back!",
		Usage:       echo.nick + "!",
	}
}
func (echo *Echo) OnEvent(event interface{}) {
	switch event.(type) {
	case oodle.Join:
		fmt.Printf("ejoin: %+v\n", event)
	case oodle.Leave:
		fmt.Printf("eleave: %+v\n", event)
	case oodle.Message:
		fmt.Printf("emsg: %+v\n", event)
	}
}

func (echo *Echo) Execute(nick string, args []string) (string, error) {
	return nick + "!", nil
}
