package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	flag "github.com/ogier/pflag"
	"github.com/oodnet/oodle/bot"
	"github.com/oodnet/oodle/discord"
	"github.com/oodnet/oodle/irc"
	"github.com/oodnet/oodle/oodle"
	"github.com/oodnet/oodle/plugins/core"
	"github.com/oodnet/oodle/plugins/hackterm"
	"github.com/oodnet/oodle/plugins/lazy"
	"github.com/oodnet/oodle/plugins/oodnet"
	"github.com/oodnet/oodle/plugins/sed"
	"github.com/oodnet/oodle/plugins/urban"
	"github.com/oodnet/oodle/plugins/webhook"
	"github.com/oodnet/oodle/plugins/wiki"
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
	go ircClient.MiniClient()
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
		urban.Register(deps),
		wiki.Register(deps),
		sed.Register(deps),
		webhook.Register(deps),
		oodnet.Register(deps),
		lazy.Register(deps),
		discord.Register(deps),
	); err != nil {
		logger.Fatal(err)
	}

	logger.Fatal(oodleBot.Start())
}
