package main

import (
	"github.com/godwhoa/oodle/bot"
	"github.com/godwhoa/oodle/plugins"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}
	oodle_bot := bot.NewBot(logger)
	plugins.RegisterEcho("oodle-dev", oodle_bot)
	err := oodle_bot.Run(bot.Config{
		Nick:    "oodle-dev",
		Server:  "chat.freenode.net",
		Port:    6697,
		Channel: "##oodle-test",
	})
	logger.Fatal(err)
}
