package oodle

import (
	"database/sql"
	"errors"

	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

var ErrUsage = errors.New("Wrong command usage")

type Command struct {
	Prefix      string
	Name        string
	Description string
	Usage       string
	Fn          func(nick string, args []string) (reply string, err error)
}

// Trigger is a basic event handler
type Trigger func(event girc.Event)

// IRCClient is a simplified client given to plugins.
type IRCClient interface {
	Connect() error
	Close()
	OnEvent(callback func(event girc.Event))
	Sender
	Checker
}

type Checker interface {
	IsRegistered(nick string) bool
	InChannel(nick string) bool
	IsAdmin(nick string) bool
}

type Sender interface {
	Send(message string)
	Sendnl(message string)
	SendTo(user, message string)
	Sendf(format string, a ...interface{})
}

// Bot handles trigger/command execution
type Bot interface {
	RegisterCommands(plugins ...Command)
	RegisterTriggers(plugins ...Trigger)
	Commands() []Command
}

// Deps is a container for common dependencies
type Deps struct {
	IRC    IRCClient
	Bot    Bot
	DB     *sql.DB
	Logger *logrus.Logger
}
