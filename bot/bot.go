package bot

import (
	"strings"

	"github.com/godwhoa/oodle/oodle"
	"github.com/jinzhu/gorm"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	triggers   []oodle.Trigger
	commandMap map[string]oodle.Command
	sendQueue  chan string
	client     *girc.Client
	log        *logrus.Logger
	config     *oodle.Config
	ircClient  *IRCClient
	db         *gorm.DB
}

func NewBot(logger *logrus.Logger, config *oodle.Config, ircClient *IRCClient, db *gorm.DB) *Bot {
	return &Bot{
		log:        logger,
		config:     config,
		ircClient:  ircClient,
		db:         db,
		commandMap: make(map[string]oodle.Command),
		sendQueue:  make(chan string, 200),
	}
}

// Start makes a conn., stats a readloop and uses the config
func (bot *Bot) Start() error {
	bot.log.Info("Connecting...")
	go bot.sendLoop()
	bot.ircClient.OnEvent(func(event interface{}) {
		if msg, ok := event.(oodle.Message); ok {
			bot.handleCommand(msg.Nick, msg.Msg)
		}
	})
	bot.ircClient.OnEvent(bot.relayTrigger)
	return bot.ircClient.Connect()
}

// Stop stops the bot in a graceful manner
func (bot *Bot) Stop() {
	bot.ircClient.Close()
}

// probably remove this in future
// keeping it for now to avoid plugin re-write
func (bot *Bot) sendLoop() {
	for message := range bot.sendQueue {
		bot.ircClient.Send(message)
	}
}

func (bot *Bot) relayTrigger(event interface{}) {
	for _, trigger := range bot.triggers {
		trigger.OnEvent(event)
	}
}

func (bot *Bot) handleCommand(nick string, message string) {
	args := strings.Split(message, " ")
	if len(args) < 1 {
		return
	}
	command, ok := bot.commandMap[args[0]]
	if !ok {
		return
	}

	reply, err := command.Execute(nick, args[1:])
	switch err {
	case oodle.ErrUsage:
		bot.sendQueue <- "Usage: " + command.Info().Usage
	case nil:
		bot.sendQueue <- reply
	default:
		bot.log.Error(err)
	}

	bot.log.WithFields(logrus.Fields{
		"cmd":    args[0],
		"caller": nick,
		"reply":  reply,
		"err":    err,
	}).Debug("CommandExec")
}

func (bot *Bot) RegisterTrigger(trigger oodle.Trigger) {
	bot.triggers = append(bot.triggers, trigger)
	trigger.SetSendQueue(bot.sendQueue)
}

func (bot *Bot) RegisterCommand(command oodle.Command) {
	cmdinfo := command.Info()
	bot.commandMap[cmdinfo.Prefix+cmdinfo.Name] = command
}

func (bot *Bot) Register(plugins ...interface{}) {
	for _, plugin := range plugins {
		if command, ok := plugin.(oodle.Command); ok {
			bot.RegisterCommand(command)
		}
		if trigger, ok := plugin.(oodle.Trigger); ok {
			bot.RegisterTrigger(trigger)
		}
		if stateful, ok := plugin.(oodle.Stateful); ok {
			stateful.Init(bot.config, bot.db)
		}
	}
}
