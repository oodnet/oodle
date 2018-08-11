package oodle

import (
	"database/sql"
	"errors"

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
type Trigger func(event interface{})

// IRCClient is a simplified client given to plugins.
type IRCClient interface {
	IsRegistered(nick string) bool
	InChannel(nick string) bool
	Send(message string)
	Sendf(format string, a ...interface{})
}

// Bot handles trigger/command execution
type Bot interface {
	// Register registers Trigger/Command only.
	Register(plugins ...interface{})
	Commands() []Command
}

// Deps is a container for common dependencies
type Deps struct {
	IRC    IRCClient
	Bot    Bot
	DB     *sql.DB
	Config *Config
	Logger *logrus.Logger
}
