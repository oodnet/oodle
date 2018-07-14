package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
	"mvdan.cc/xurls"
)

// Register wires everything up
func Register(deps *oodle.Deps) error {
	irc, bot, config, db := deps.IRC, deps.Bot, deps.Config, deps.DB
	seenCmd, seenTrig := Seen()
	tellCmd, tellTrig := Tell(irc, db)
	bot.Register(
		Echo(config.Nick),
		CustomCommands(config.CustomCommands, irc),
		TitleScraper(irc),
		seenCmd, seenTrig,
		tellCmd, tellTrig,
	)

	return nil
}

// Echo echoes back your nick!
func Echo(botNick string) oodle.Command {
	return oodle.Command{
		Prefix:      "",
		Name:        botNick + "!",
		Description: "Exlamates your nick back!",
		Usage:       botNick + "!",
		Fn: func(nick string, args []string) (string, error) {
			return nick + "!", nil
		},
	}
}

// CustomCommands lets you make custom commands via. config
func CustomCommands(commands map[string]string, irc oodle.IRCClient) oodle.Trigger {
	return func(event interface{}) {
		message, ok := event.(oodle.Message)
		if !ok {
			return
		}
		args := strings.Split(strings.TrimSpace(message.Msg), " ")
		if len(args) < 1 {
			return
		}
		if msg, ok := commands[args[0]]; ok {
			irc.Send(msg)
		}
	}
}

// Seen tells you when it last saw someone
func Seen() (oodle.Command, oodle.Trigger) {
	store := make(map[string]time.Time)
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "seen",
		Description: "Tells you when it last saw someone",
		Usage:       ".seen <nick>",
		Fn: func(nick string, args []string) (string, error) {
			if len(args) < 1 {
				return "", oodle.ErrUsage
			}
			if lastSeen, ok := store[args[0]]; ok {
				return fmt.Sprintf("%s was last seen %s ago.", args[0], u.FmtTime(lastSeen)), nil
			}
			return fmt.Sprintf("No logs for %s", args[0]), nil
		},
	}
	trigger := func(event interface{}) {
		switch event.(type) {
		case oodle.Join:
			store[event.(oodle.Join).Nick] = time.Now()
		case oodle.Leave:
			store[event.(oodle.Leave).Nick] = time.Now()
		case oodle.Message:
			store[event.(oodle.Message).Nick] = time.Now()
		}
	}
	return cmd, trigger
}

// TitleScraper fetches, scrapes and sends titles whenever it sees urls
func TitleScraper(irc oodle.IRCClient) oodle.Trigger {
	urlReg := xurls.Strict

	return func(event interface{}) {
		message, ok := event.(oodle.Message)
		if !ok {
			return
		}
		urls := urlReg.FindAllString(message.Msg, -1)
		for _, url := range urls {
			if url != "" {
				doc, err := goquery.NewDocument(url)
				if err != nil {
					return
				}
				pageTitle := doc.Find("title").First().Text()
				irc.Send(strings.TrimSpace(pageTitle))
			}
		}
	}
}
