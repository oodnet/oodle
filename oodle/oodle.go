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

// Trigger can listen for a event and get triggered and send messages via. the send queue
type Trigger interface {
	SetSendQueue(sendqueue chan string)
	OnEvent(event interface{})
}

type Stateful interface {
	Init(config *Config, db *gorm.DB)
}

type Bot interface {
	RegisterTrigger(trigger Trigger)
	RegisterCommand(command Command)
}
