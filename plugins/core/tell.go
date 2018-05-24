package core

import (
	"database/sql"
	"strings"
	"time"

	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
	"github.com/jmoiron/sqlx"
)

// Tell lets users send a msg. to an inactive user
func Tell(irc oodle.IRCClient, db *sql.DB) (oodle.Command, oodle.Trigger) {
	store := NewTellStore(sqlx.NewDb(db, "sqlite3"))

	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "tell",
		Description: "Lets you send a msg. to an inactive user.\nIt will notify them once they are active.",
		Usage:       ".tell <to> <msg>",
		Fn: func(nick string, args []string) (string, error) {
			if len(args) < 2 {
				return "", oodle.ErrUsage
			}
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
	notify := func(nick string) {
		if !store.HasMail(nick) {
			return
		}
		letters := store.GetLetters(nick)
		for _, l := range letters {
			irc.Sendf("%s, %s left this message for you: %s\n%s ago", nick, l.From, l.Body, u.FmtTime(l.When))
		}
		store.Delete(letters)
	}
	trigger := func(event interface{}) {
		switch event.(type) {
		case oodle.Join:
			notify(event.(oodle.Join).Nick)
		case oodle.Message:
			notify(event.(oodle.Message).Nick)
		}
	}
	return cmd, trigger
}
