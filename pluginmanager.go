package main

import (
	"database/sql"

	"github.com/godwhoa/oodle/oodle"
	_ "github.com/mattn/go-sqlite3"
)

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if needle == h {
			return true
		}
	}
	return false
}

type PluginManager struct {
	db     *sql.DB
	irc    oodle.IRCClient
	config *oodle.Config
	bot    oodle.Bot
}

func NewPM(irc oodle.IRCClient, config *oodle.Config, bot oodle.Bot) (*PluginManager, error) {
	db, err := sql.Open("sqlite3", config.DBPath)
	if err != nil {
		return nil, err
	}
	return &PluginManager{
		db:     db,
		irc:    irc,
		config: config,
		bot:    bot,
	}, nil
}

func (pm *PluginManager) RegisterPlugin(name string, plugin interface{}) {
	if !contains(pm.config.Commands, name) {
		return
	}

	bot, irc, config, db := pm.bot, pm.irc, pm.config, pm.db

	if interactive, ok := plugin.(oodle.Interactive); ok {
		interactive.SetIRC(irc)
	}
	if configureable, ok := plugin.(oodle.Configureable); ok {
		configureable.Configure(config)
	}
	if persistable, ok := plugin.(oodle.Persistable); ok {
		persistable.SetDB(db)
	}

	if command, ok := plugin.(oodle.Command); ok {
		bot.RegisterCommand(command)
	}
	if trigger, ok := plugin.(oodle.Trigger); ok {
		bot.RegisterTrigger(trigger)
	}
}
