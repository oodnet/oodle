package plugins

import (
	"database/sql"
	"strings"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/store"
	"github.com/hako/durafmt"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
)

func fmtTime(t time.Time) string {
	// gets rid of milliseconds, I think?
	since := time.Since(t).Truncate(time.Second)
	// formats it to 1 day etc.
	return durafmt.Parse(since).String()
}

type Letter struct {
	gorm.Model
	From string
	To   string
	Body string
	When time.Time
}

type Tell struct {
	store *store.TellStore
	oodle.BaseInteractive
}

func (t *Tell) notify(nick string) {
	if !t.store.HasMail(nick) {
		return
	}
	letters := t.store.GetLetters(nick)
	for _, l := range letters {
		t.IRC.Sendf("%s, %s left this message for you: %s\n%s ago", nick, l.From, l.Body, fmtTime(l.When))
	}
	t.store.Delete(letters)
}

func (t *Tell) SetDB(db *sql.DB) {
	t.store = store.NewTellStore(sqlx.NewDb(db, "sqlite3"))
}

func (tell *Tell) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "tell",
		Description: "Lets you send a msg. to an inactive user.\nIt will notify them once they are active.",
		Usage:       ".tell <to> <msg>",
	}
}

func (t *Tell) OnEvent(event interface{}) {
	switch event.(type) {
	case oodle.Join:
		t.notify(event.(oodle.Join).Nick)
	case oodle.Message:
		t.notify(event.(oodle.Message).Nick)
	}
}

func (t *Tell) Execute(nick string, args []string) (string, error) {
	if len(args) < 2 {
		return "", oodle.ErrUsage
	}
	err := t.store.Send(store.Letter{
		From: nick,
		To:   args[0],
		Body: strings.Join(args[1:], " "),
		When: time.Now(),
	})
	if err != nil {
		return "Internal Error", nil
	}
	return "okie dokie!", nil
}
