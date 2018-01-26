package main

import (
	"github.com/godwhoa/oodle/bot"
	"github.com/godwhoa/oodle/plugins"
)

func main() {
	oodle_bot := bot.NewBot()
	plugins.RegisterEcho("oodle-dev", oodle_bot)
	oodle_bot.Run(bot.Config{
		Nick:    "oodle-dev",
		Server:  "chat.freenode.net",
		Port:    6697,
		Channel: "##oodle-test",
	})
}
