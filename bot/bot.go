package bot

import (
	"strings"

	"github.com/lrstanley/girc"
	"github.com/oodnet/oodle/events"
	"github.com/oodnet/oodle/oodle"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Bot struct {
	triggers   []oodle.Trigger
	commandMap map[string]oodle.Command
	log        *logrus.Logger
	irc        oodle.IRCClient
}

func NewBot(logger *logrus.Logger, ircClient oodle.IRCClient) *Bot {
	return &Bot{
		log:        logger,
		irc:        ircClient,
		commandMap: make(map[string]oodle.Command),
	}
}

// Start makes a conn., stats a readloop and uses the config
func (bot *Bot) Start() error {
	bot.log.Info("Connecting...")
	bot.irc.OnEvent(func(event girc.Event) {
		if events.Is(event, events.MESSAGE) {
			bot.handleCommand(events.Message(event))
		}
	})
	bot.irc.OnEvent(bot.relayTrigger)
	return bot.irc.Connect()
}

// Stop stops the bot in a graceful manner
func (bot *Bot) Stop() {
	bot.irc.Close()
}

func (bot *Bot) relayTrigger(event girc.Event) {
	for _, trigger := range bot.triggers {
		trigger(event)
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

	reply, err := command.Fn(nick, args[1:])
	switch err {
	case oodle.ErrUsage:
		bot.irc.Sendf("Usage: " + command.Usage)
	case nil:
		bot.irc.Sendnl(reply)
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (bot *Bot) RegisterCommands(commands ...oodle.Command) {
	disabled := viper.GetStringSlice("disabled_commands")
	for _, cmd := range commands {
		if contains(disabled, cmd.Name) {
			return
		}
		bot.commandMap[cmd.Prefix+cmd.Name] = cmd
	}
}

func (bot *Bot) RegisterTriggers(triggers ...oodle.Trigger) {
	bot.triggers = append(bot.triggers, triggers...)
}

func (bot *Bot) Commands() []oodle.Command {
	cmds := []oodle.Command{}
	for _, cmd := range bot.commandMap {
		cmds = append(cmds, cmd)
	}
	return cmds
}
