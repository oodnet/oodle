package bot

import (
	"fmt"
	"strings"

	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
)

type Config struct {
	Nick    string
	Server  string
	Port    int
	Channel string
}

type Bot struct {
	triggers   []oodle.Trigger
	commandmap map[string]oodle.Command
	sendQueue  chan string
	client     *girc.Client
}

func NewBot() *Bot {
	return &Bot{
		commandmap: make(map[string]oodle.Command),
		sendQueue:  make(chan string, 200),
	}
}

func (bot *Bot) Run(config Config) error {
	// TODO: refactor this
	// Also use context
	// girc lets you specify recovery function
	client := girc.New(girc.Config{
		Server: config.Server,
		Port:   6667,
		Nick:   config.Nick,
		User:   config.Nick + "_user",
		Name:   config.Nick + "_name",
		// Debug:  os.Stdout,
	})

	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
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
	return bot.client.Connect()
}

func (bot *Bot) sendLoop(client *girc.Client, channel string) {
	for message := range bot.sendQueue {
		for _, msg := range strings.Split(message, "\n") {
			client.Cmd.Message(channel, msg)
		}
	}
}

func (bot *Bot) sendEvent(event interface{}) {
	fmt.Println(event)
	for _, trigger := range bot.triggers {
		trigger.OnEvent(event)
	}
}

func (bot *Bot) handleCommand(nick string, message string) {
	args := strings.Split(message, " ")
	if len(args) < 1 {
		return
	}
	command, ok := bot.commandmap[args[0]]
	if !ok {
		return
	}
	reply, err := command.Execute(nick, args[1:])
	switch err {
	case oodle.ErrUsage:
		bot.sendQueue <- command.Info().Usage
	case nil:
		bot.sendQueue <- reply
	default:
		fmt.Println(err)
	}
}

func (bot *Bot) RegisterTrigger(trigger oodle.Trigger) {
	bot.triggers = append(bot.triggers, trigger)
	trigger.SetSendQueue(bot.sendQueue)
}

func (bot *Bot) RegisterCommand(command oodle.Command) {
	cmdinfo := command.Info()
	bot.commandmap[cmdinfo.Prefix+cmdinfo.Name] = command
}
