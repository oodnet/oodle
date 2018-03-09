package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/godwhoa/oodle/oodle"
)

func contains(haystack []string, needle string) bool {
	for _, h := haystack{
		if needle == h{
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

func NewPM(irc oodle.IRCClient, config *oodle.Config, bot oodle.Bot)  (*PluginManager, error){
	db, err := sql.Open("sqlite3", config.DBPath)
	if err != nil{
		return err
	}
	return &PluginManager{
		db: db,
		irc:irc,
		config:config,
		bot:bot,
	}
}

func (pm *PluginManager) RegisterPlugin(name string, plugin interface{}) {
	if !contains(pm.config, name){
		return
	}
	bot, irc, config, db := pm.bot, pm.irc, pm.config, pm.db
	if command, ok := plugin.(oodle.Command); ok {
		bot.RegisterCommand(command)
	}
	if trigger, ok := plugin.(oodle.Trigger); ok {
		bot.RegisterTrigger(trigger)
	}
	if interactive, ok := plugin.(oodle.Interactive); ok {
		inta.SetIRC(irc)
	}
	if configureable, ok := plugin.(oodle.Configureable); ok {
		configureable.Configure(config)
	}
	if persistable, ok := plugin.(oodle.Persistable); ok {
		persistable.SetDB(db)
	}
}
