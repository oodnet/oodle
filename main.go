package main

import (
	"database/sql"
	"os"

	"github.com/godwhoa/oodle/bot"
	"github.com/godwhoa/oodle/irc"
	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/plugins/core"
	"github.com/godwhoa/oodle/plugins/hackterm"
	"github.com/godwhoa/oodle/plugins/webhook"
	_ "github.com/mattn/go-sqlite3"
	flag "github.com/ogier/pflag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func must(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	confpath := flag.StringP("config", "c", "config.toml", "Specifies which configfile to use")
	doUpgrade := flag.BoolP("upgrade", "u", false, "Upgrades oodle")
	flag.Parse()

	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{DisableColors: true}
	logger.SetLevel(logrus.DebugLevel)

	if *doUpgrade {
		if err := upgrade(); err != nil {
			logger.Fatal(err)
		}
		os.Exit(0)
	}

	viper.SetConfigFile(*confpath)
	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal(err)
	}

	oodle.SetDefaults()

	db, err := sql.Open("sqlite3", viper.GetString("dbpath"))
	if err != nil {
		logger.Fatal(err)
	}

	ircClient := irc.NewIRCClient(logger)
	oodleBot := bot.NewBot(logger, ircClient)

	deps := &oodle.Deps{
		IRC:    ircClient,
		Bot:    oodleBot,
		Logger: logger,
		DB:     db,
	}

	// register plugins and log failure
	if err := must(
		core.Register(deps),
		hackterm.Register(deps),
		webhook.Register(deps),
	); err != nil {
		logger.Fatal(err)
	}

	logger.Fatal(oodleBot.Start())
}
