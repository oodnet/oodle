package main

import (
	"github.com/BurntSushi/toml"
	"github.com/godwhoa/oodle/bot"
	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/plugins"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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

	db, err := gorm.Open("sqlite3", config.DBPath)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	ircClient := bot.NewIRCClient(logger, config)
	webhook := NewWebHook(ircClient, logger, config.Secret)
	commandMap := map[string]interface{}{
		"seen":     &plugins.Seen{},
		"tell":     &plugins.Tell{},
		"echo":     &plugins.Echo{Nick: config.Nick},
		"title":    &plugins.Title{},
		"give":     &plugins.Give{},
		"rank":     &plugins.Rank{},
		"hackterm": &plugins.HackTerm{},
	}

	oodleBot := bot.NewBot(logger, config, ircClient, db)
	for _, commandName := range config.Commands {
		if command, ok := commandMap[commandName]; ok {
			oodleBot.Register(command)
		}
	}

	go webhook.Listen(config.WebHookAddr)
	logger.Fatal(oodleBot.Start())
}
