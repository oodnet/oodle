package core

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/PuerkitoBio/goquery"
	m "github.com/godwhoa/oodle/middleware"
	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"mvdan.cc/xurls"
)

func ａs(text string) string {
	var builder strings.Builder
	for _, r := range text {
		offset := rune(0)
		if r >= 33 && r <= 126 {
			offset = 65248
		}
		builder.WriteRune(r + offset)
	}
	return builder.String()
}

// Register wires everything up
func Register(deps *oodle.Deps) error {
	irc, bot, db := deps.IRC, deps.Bot, deps.DB
	remindin := &RemindIn{irc, irc, NewReminderStore(db), NewMailBox(db)}
	go remindin.Watch()
	seen := &Seen{db: sqlx.NewDb(db, "sqlite3")}
	seen.Migrate()
	tellCmd, tellTrig := Tell(irc, db)
	bot.RegisterTriggers(
		CustomCommands(irc),
		ExecCommands(irc),
		TitleScraper(irc),
		seen.Trigger(),
		tellTrig,
	)
	bot.RegisterCommands(
		Echo(),
		Version(),
		GC(irc),
		Memory(irc),
		List(bot, irc),
		Help(bot),
		remindin.Command(),
		seen.Command(),
		tellCmd,
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

func GC(checker oodle.Checker) oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "gc",
		Description: "Runs GC; debug purpose only.",
		Usage:       ".gc",
		Fn: func(nick string, args []string) (reply string, err error) {
			runtime.GC()
			debug.FreeOSMemory()
			return "Ran runtime.GC() and debug.FreeOSMemory()", nil
		},
	}
	return m.Chain(cmd, m.AdminOnly(checker))
}

func Memory(checker oodle.Checker) oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "mem",
		Description: "Shows memory usage",
		Usage:       ".mem",
		Fn: func(nick string, args []string) (reply string, err error) {
			var m runtime.MemStats
			var gc debug.GCStats
			runtime.ReadMemStats(&m)
			debug.ReadGCStats(&gc)
			alloc := m.Alloc / 1024 / 1024
			talloc := m.TotalAlloc / 1024 / 1024
			sys := m.Sys / 1024 / 1024
			return fmt.Sprintf("Alloc: %d MiB TotalAlloc: %d MiB Sys: %d MiB LastGC: %s; https://godoc.org/runtime#MemStats", alloc, talloc, sys, u.FmtTime(gc.LastGC)), nil
		},
	}
	return m.Chain(cmd, m.AdminOnly(checker))
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
		Description: "List of all the commands",
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
		Description: "Ｅｘｃｌａｍａｔｅｓ　ｙｏｕｒ　ｎｉｃｋ　ｂａｃｋ！",
		Usage:       botNick + "!",
		Fn: func(nick string, args []string) (string, error) {
			return ａs(nick + "!"), nil
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

// ExecCommands is similar to custom commands but instead of sending fixed message
// It runs/sends output of configured shell command
func ExecCommands(sender oodle.Sender) oodle.Trigger {
	commands := viper.GetStringMapString("exec_commands")
	return func(event interface{}) {
		message, ok := event.(oodle.Message)
		if !ok {
			return
		}
		args := strings.Split(strings.TrimSpace(message.Msg), " ")
		if len(args) < 1 {
			return
		}
		if command, ok := commands[args[0]]; ok {
			go func() {
				cmd := exec.Command("bash", "-c", command)
				output, _ := cmd.CombinedOutput()
				sender.Send(string(output))
			}()
		}
	}
}

func newDocument(url string) (*goquery.Document, error) {
	// Load the URL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Oodlebot/1.0")

	response, err := u.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	response.Body = u.LimitBody(response.Body, 1)
	return goquery.NewDocumentFromResponse(response)
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
