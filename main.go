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
	logger.SetLevel(logrus.DebugLevel)

	config := &oodle.Config{}
	if _, err := toml.DecodeFile("config.toml", config); err != nil {
		logger.Fatal(err)
	}

	oodleBot := bot.NewBot(logger)
	// TODO: better way of registering plugins
	plugins.RegisterEcho(config, oodleBot)
	plugins.RegisterSeen(config, oodleBot)

	logger.Fatal(oodleBot.Run(config))
}
