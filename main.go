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

	config := &oodle.Config{}
	if _, err := toml.DecodeFile("config.toml", config); err != nil {
		logger.Fatal(err)
	}

	oodleBot := bot.NewBot(logger)
	plugins.RegisterEcho(config, oodleBot)

	logger.Fatal(oodleBot.Run(config))
}
