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
	}
}

// Start makes a conn., stats a readloop and uses the config
func (bot *Bot) Start() error {
	bot.log.Info("Connecting...")
	bot.ircClient.OnEvent(func(event interface{}) {
		if msg, ok := event.(oodle.Message); ok {
			go bot.handleCommand(msg.Nick, msg.Msg)
		}
	})
	bot.ircClient.OnEvent(bot.relayTrigger)
	return bot.ircClient.Connect()
}

// Stop stops the bot in a graceful manner
func (bot *Bot) Stop() {
	bot.ircClient.Close()
}

func (bot *Bot) relayTrigger(event interface{}) {
	for _, trigger := range bot.triggers {
		go trigger.OnEvent(event)
	}
}

func (bot *Bot) handleCommand(nick string, message string) {
	args := strings.Split(strings.TrimSpace(message), " ")
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
		bot.ircClient.Sendf("Usage: " + command.Info().Usage)
	case nil:
		bot.ircClient.Send(reply)
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
}

func (bot *Bot) RegisterCommand(command oodle.Command) {
	cmdinfo := command.Info()
	bot.commandMap[cmdinfo.Prefix+cmdinfo.Name] = command
}
