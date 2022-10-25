package core

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lrstanley/girc"
	"github.com/oodnet/oodle/events"
	m "github.com/oodnet/oodle/middleware"
	"github.com/oodnet/oodle/oodle"
	u "github.com/oodnet/oodle/utils"
)

// Tell lets users send a msg. to an inactive user
func Tell(irc oodle.IRCClient, db *sql.DB) ([]oodle.Command, oodle.Trigger) {
	store := NewMailBox(db)

	tell := oodle.Command{
		Prefix:      ".",
		Name:        "tell",
		Description: "Lets you send a msg. to an inactive user. It will notify them once they are active.",
		Usage:       ".tell <to> <msg>",
		Fn: func(nick string, args []string) (string, error) {
			err := store.Send(Letter{
				From: nick,
				To:   args[0],
				Body: strings.Join(args[1:], " "),
				When: time.Now(),
			})
			if err != nil {
				return "Internal Error", nil
			}
			return "okie dokie!", nil
		},
	}
	unsend := oodle.Command{
		Prefix:      ".",
		Name:        "unsend",
		Description: "Cancels the last .tell message from you which is still pending.",
		Usage:       `.unsend`,
		Fn: func(nick string, args []string) (string, error) {
			letters := store.LettersFrom(nick)
			if len(letters) < 1 {
				return "No pending msgs. to cancel.", nil
			}
			store.Delete(letters[0].ID)
			return fmt.Sprintf("Your msg. to %s cancelled.", letters[0].To), nil
		},
	}

	pending := oodle.Command{
		Prefix:      ".",
		Name:        "pending",
		Description: "Shows pending .tell messages sent by you.",
		Usage:       `.pending`,
		Fn: func(nick string, args []string) (string, error) {
			letters := store.LettersFrom(nick)
			if len(letters) < 1 {
				return "No pending messages.", nil
			}
			output := ""
			for _, letter := range letters {
				limit := u.Min(20, len(letter.Body))
				output += fmt.Sprintf("ID: %d To: %s Msg: %s...\n", letter.ID, letter.To, letter.Body[0:limit])
			}
			return strings.TrimSuffix(output, "\n"), nil
		},
	}

	cancel := oodle.Command{
		Prefix:      ".",
		Name:        "cancel",
		Description: "Cancels the last .tell msg from you which is still pending.",
		Usage:       `.cancel <id>`,
		Fn: func(nick string, args []string) (string, error) {
			letters := store.LettersFrom(nick)
			if len(letters) < 1 {
				return "No pending msgs. to cancel.", nil
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return "Invalid id", nil
			}

			letter, err := store.Get(uint(id))
			if err != nil {
				return "No msg. with id " + args[0], nil
			}

			if letter.From != nick {
				return "No msg. with id " + args[0] + " belonging to you.", nil
			}
			store.Delete(uint(id))
			return fmt.Sprintf("Your msg. to %s cancelled.", letters[0].To), nil
		},
	}

	notify := func(nick string) {
		if !store.HasMail(nick) {
			return
		}
		letters := store.LettersTo(nick)
		for _, l := range letters {
			irc.Sendf("%s, %s left this message for you: %s\n%s ago", nick, l.From, l.Body, u.FmtTime(l.When))
		}
		store.BatchDelete(letters)
	}
	trigger := func(event girc.Event) {
		if events.Is(event, events.MESSAGE) {
			notify(events.Nick(event))
		}
	}
	tell = m.Chain(tell, m.MinArg(2))
	cancel = m.Chain(cancel, m.MinArg(1))
	return []oodle.Command{tell, cancel, unsend, pending}, trigger
}
