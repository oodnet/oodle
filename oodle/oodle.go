package oodle

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var ErrUsage = errors.New("Wrong command usage")

// CommandInfo contains information of a command
type CommandInfo struct {
	Prefix      string
	Name        string
	Description string
	Usage       string
}

// Command is like a pure function just inputs and outputs
type Command interface {
	Info() CommandInfo
	Execute(nick string, args []string) (reply string, err error)
}

// IRCClient is a simplified client given to plugins.
type IRCClient interface {
	IsRegistered(nick string) bool
	InChannel(nick string) bool
	Send(message string)
	Sendf(format string, a ...interface{})
}

// Interactive for plugins that interact with irc
type Interactive interface {
	SetIRC(irc IRCClient)
}

// Trigger can listen for a event and get triggered and send messages via. the send queue
type Trigger interface {
	OnEvent(event interface{})
}

// Persistable for plugins that need to persist
type Persistable interface {
	SetDB(db *gorm.DB)
}

// Configureable for plugins that depend on config
type Configureable interface {
	Configure(config *Config)
}

type Bot interface {
	RegisterTrigger(trigger Trigger)
	RegisterCommand(command Command)
}
