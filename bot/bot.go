package bot

import (
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Nick    string
	Server  string
	Port    int
	Channel string
}

type Bot struct {
	triggers   []oodle.Trigger
	commandMap map[string]oodle.Command
	sendQueue  chan string
	client     *girc.Client
	log        *logrus.Logger
}

func NewBot(logger *logrus.Logger) *Bot {
	return &Bot{
		log:        logger,
		commandMap: make(map[string]oodle.Command),
		sendQueue:  make(chan string, 200),
	}
}

// Run runs the bot with a config
func (bot *Bot) Run(config *oodle.Config) error {
	gircConf := girc.Config{
		Server:      config.Server,
		Port:        config.Port,
		Nick:        config.Nick,
		User:        config.Nick + "_user",
		Name:        config.Nick + "_name",
		RecoverFunc: func(_ *girc.Client, e *girc.HandlerError) { bot.log.Errorln(e.Error()) },
	}
	if config.SASLUser != "" && config.SASLPass != "" {
		gircConf.SASL = &girc.SASLPlain{User: config.SASLUser, Pass: config.SASLPass}
	}
	client := girc.New(gircConf)

	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		bot.log.WithFields(logrus.Fields{
			"server":  config.Server,
			"port":    config.Port,
			"channel": config.Channel,
			"nick":    client.GetNick(),
		}).Info("Connected!")
		c.Cmd.Join(config.Channel)
	})

	// Channel trigger
	client.Handlers.Add(girc.JOIN, func(c *girc.Client, e girc.Event) {
		nick := e.Source.Name
		if nick != config.Nick {
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
		if nick != config.Nick {
			bot.handleCommand(nick, msg)
			bot.sendEvent(oodle.Message{Nick: nick, Msg: msg})
		}
	})

	go bot.sendLoop(client, config.Channel)

	bot.client = client
	if err, ok := client.Connect().(*girc.ErrInvalidConfig); ok {
		return err
	}
	return backoff.RetryNotify(client.Connect, backoff.NewExponentialBackOff(), func(err error, dur time.Duration) {
		bot.log.Warnf("Connection failed with err: %s\n Retrying in %s", err.Error(), dur.String())
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
