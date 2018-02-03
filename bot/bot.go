package bot

import (
	"strings"
	"time"

	"github.com/cenkalti/backoff"
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
	db         *gorm.DB
}

func NewBot(logger *logrus.Logger, config *oodle.Config, db *gorm.DB) *Bot {
	return &Bot{
		log:        logger,
		config:     config,
		db:         db,
		commandMap: make(map[string]oodle.Command),
		sendQueue:  make(chan string, 200),
	}
}

// Start makes a conn., stats a readloop and uses the config
// FIXME: bot.config is ugly, find a better way
func (bot *Bot) Start() error {
	gircConf := girc.Config{
		Server:      bot.config.Server,
		Port:        bot.config.Port,
		Nick:        bot.config.Nick,
		User:        bot.config.Nick + "_user",
		Name:        bot.config.Nick + "_name",
		RecoverFunc: func(_ *girc.Client, e *girc.HandlerError) { bot.log.Errorln(e.Error()) },
	}
	if bot.config.SASLUser != "" && bot.config.SASLPass != "" {
		gircConf.SASL = &girc.SASLPlain{User: bot.config.SASLUser, Pass: bot.config.SASLPass}
	}
	client := girc.New(gircConf)

	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		bot.log.WithFields(logrus.Fields{
			"server":  bot.config.Server,
			"port":    bot.config.Port,
			"channel": bot.config.Channel,
			"nick":    client.GetNick(),
		}).Info("Connected!")
		c.Cmd.Join(bot.config.Channel)
	})

	// Channel trigger
	client.Handlers.Add(girc.JOIN, func(c *girc.Client, e girc.Event) {
		nick := e.Source.Name
		if nick != bot.config.Nick {
			bot.sendEvent(oodle.Join{Nick: nick})
		}
	})

	client.Handlers.Add(girc.PART, func(c *girc.Client, e girc.Event) {
		nick := e.Source.Name
		bot.sendEvent(oodle.Leave{Nick: nick})
	})

	client.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		nick := e.Source.Name
		msg := e.Trailing
		if nick != bot.config.Nick {
			bot.handleCommand(nick, msg)
			bot.sendEvent(oodle.Message{Nick: nick, Msg: msg})
		}
	})

	bot.log.Info("Connecting...")
	go bot.sendLoop(client, bot.config.Channel)

	bot.client = client
	err := client.Connect()
	if _, ok := err.(*girc.ErrInvalidConfig); ok || !bot.config.Retry {
		return err
	}
	return backoff.RetryNotify(client.Connect, backoff.NewExponentialBackOff(), func(err error, dur time.Duration) {
		bot.log.Warnf("Connection failed with err: %s", err)
		bot.log.Warnf("Retrying in %s", dur)
	})
}

// Stop stops the bot in a graceful manner
func (bot *Bot) Stop() {
	bot.client.Close()
}

func (bot *Bot) sendLoop(client *girc.Client, channel string) {
	for message := range bot.sendQueue {
		for _, msg := range strings.Split(message, "\n") {
			client.Cmd.Message(channel, msg)
		}
	}
}

func (bot *Bot) sendEvent(event interface{}) {
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
