package core

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/PuerkitoBio/goquery"
	m "github.com/godwhoa/oodle/middleware"
	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
	"mvdan.cc/xurls"
)

// Register wires everything up
func Register(deps *oodle.Deps) error {
	irc, bot, db := deps.IRC, deps.Bot, deps.DB
	remindin := &RemindIn{irc, irc, NewReminderStore(db), NewMailBox(db)}
	go remindin.Watch()
	seenCmd, seenTrig := Seen()
	tellCmd, tellTrig := Tell(irc, db)
	bot.Register(
		Echo(),
		CustomCommands(irc),
		TitleScraper(irc),
		Version(),
		List(bot, irc),
		Help(bot),
		remindin.Command(),
		seenCmd, seenTrig,
		tellCmd, tellTrig,
	)

	return nil
}

func Version() oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "version",
		Description: "Shows version and commit",
		Usage:       ".version",
		Fn: func(nick string, args []string) (reply string, err error) {
			return fmt.Sprintf("Version: %s Commit: %s", oodle.Version, oodle.Commit), nil
		},
	}
}

// Help gives help info for commands
func Help(bot oodle.Bot) oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "help",
		Description: "Give description and usage for a command",
		Usage:       ".help <command name>",
		Fn: func(nick string, args []string) (reply string, err error) {
			for _, cmd := range bot.Commands() {
				if cmd.Name == args[0] {
					return fmt.Sprintf("Desciption: %s\nUsage: %s", cmd.Description, cmd.Usage), nil
				}
			}
			return "No command named " + args[0], nil
		},
	}
	cmd = m.Chain(cmd, m.MinArg(1))
	return cmd
}

// List lists all the commands
func List(bot oodle.Bot, sender oodle.Sender) oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "list",
		Description: "PMs you a list of all the commands",
		Usage:       ".list",
		Fn: func(nick string, args []string) (reply string, err error) {
			msg := "Commands: "
			for _, cmd := range bot.Commands() {
				msg += cmd.Name
				msg += ", "
			}
			msg = strings.TrimSuffix(msg, ", ")
			return msg, nil
		},
	}
}

// Echo echoes back your nick!
func Echo() oodle.Command {
	botNick := viper.GetString("nick")
	return oodle.Command{
		Prefix:      "",
		Name:        botNick + "!",
		Description: "Exclamates your nick back!",
		Usage:       botNick + "!",
		Fn: func(nick string, args []string) (string, error) {
			return nick + "!", nil
		},
	}
}

// CustomCommands lets you make custom commands via. config
func CustomCommands(irc oodle.IRCClient) oodle.Trigger {
	commands := viper.GetStringMapString("custom_commands")
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
	cmd = m.Chain(cmd, m.MinArg(1))
	return cmd, trigger
}

func newDocument(url string) (*goquery.Document, error) {
	// Load the URL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Oodlebot/1.0")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromResponse(res)
}

func checker() func(url string) bool {
	blacklist := viper.GetStringSlice("url_blacklist")
	return func(url string) bool {
		for _, burl := range blacklist {
			if strings.Contains(url, burl) {
				return true
			}
		}
		return false
	}
}

// TitleScraper fetches, scrapes and sends titles whenever it sees urls
func TitleScraper(irc oodle.IRCClient) oodle.Trigger {
	urlReg := xurls.Strict
	isblacklisted := checker()

	return func(event interface{}) {
		message, ok := event.(oodle.Message)
		if !ok {
			return
		}
		urls := urlReg.FindAllString(message.Msg, -1)
		for _, url := range urls {
			if url != "" && !isblacklisted(url) {
				doc, err := newDocument(url)
				if err != nil {
					return
				}
				pageTitle := doc.Find("title").First().Text()
				irc.Send(strings.TrimSpace(pageTitle))
			}
		}
	}
}
