package main

import (
	"github.com/BurntSushi/toml"
	"github.com/godwhoa/oodle/bot"
	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/plugins"
	flag "github.com/ogier/pflag"
	"github.com/sirupsen/logrus"
)

func main() {
	confpath := *flag.StringP("config", "c", "config.toml", "Specifies which configfile to use")
	flag.Parse()

	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}
	logger.SetLevel(logrus.DebugLevel)

	config := &oodle.Config{}
	if _, err := toml.DecodeFile(confpath, config); err != nil {
		logger.Fatal(err)
	}
	if len(config.Cooldowns) != len(config.Points) {
		logger.Fatal("config: len(cooldowns) != len(points)")
	}

	ircClient := bot.NewIRCClient(logger, config)

	oodleBot := bot.NewBot(logger, ircClient)

	pm, err := NewPM(ircClient, config, oodleBot)
	if err != nil {
		logger.Fatal(err)
	}
	defer pm.db.Close()

	pm.RegisterPlugin("seen", &plugins.Seen{})
	pm.RegisterPlugin("tell", &plugins.Tell{})
	pm.RegisterPlugin("echo", &plugins.Echo{})
	pm.RegisterPlugin("title", &plugins.Title{})
	pm.RegisterPlugin("give", &plugins.Give{})
	pm.RegisterPlugin("rep", &plugins.Rep{})
	pm.RegisterPlugin("rank", &plugins.Rank{})
	pm.RegisterPlugin("hackterm", &plugins.HackTerm{})

	webhook := NewWebHook(ircClient, logger, config.Secret)
	go webhook.Listen(config.WebHookAddr)
	logger.Fatal(oodleBot.Start())
}
