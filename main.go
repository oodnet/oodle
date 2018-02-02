package main

import (
	"github.com/BurntSushi/toml"
	"github.com/godwhoa/oodle/bot"
	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/plugins"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}

	oodleBot := bot.NewBot(logger)
	plugins.RegisterEcho("oodle-dev", oodleBot)

	config := &oodle.Config{}
	if _, err := toml.DecodeFile("config.toml", config); err != nil {
		logger.Fatal(err)
	}
	logger.Fatal(oodleBot.Run(config))
}
