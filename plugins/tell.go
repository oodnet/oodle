package plugins

import (
	"fmt"
	"strings"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/hako/durafmt"
	"github.com/jinzhu/gorm"
)

type Letter struct {
	gorm.Model
	From string
	To   string
	Body string
	When time.Time
}

type Tell struct {
	db *gorm.DB
	// cache for checking if a user has mail
	// to avoid querying for every msg.
	has map[string]bool
	oodle.BaseTrigger
}

func RegisterTell(config *oodle.Config, db *gorm.DB, bot oodle.Bot) {
	db.AutoMigrate(&Letter{})
	tell := &Tell{db: db, has: make(map[string]bool)}
	bot.RegisterCommand(tell)
	bot.RegisterTrigger(tell)
}

func (tell *Tell) notify(nick string) {
	if c, ok := tell.has[nick]; ok && !c {
		return
	}
	var letters []Letter
	tell.db.Where("`to` = ?", nick).Find(&letters)
	for _, l := range letters {
		timeSince := time.Since(l.When)
		tell.SendQueue <- fmt.Sprintf("%s, %s left this message for you: %s\n%s ago", nick, l.From, l.Body, durafmt.Parse(timeSince).String())
		tell.db.Delete(&l)
	}
	tell.has[nick] = false
}

func (tell *Tell) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "tell",
		Description: "Lets you send a msg. to an inactive user.\nIt will notify them once they are active.",
		Usage:       ".",
	}
}

func (tell *Tell) OnEvent(event interface{}) {
	switch event.(type) {
	case oodle.Join:
		tell.notify(event.(oodle.Join).Nick)
	case oodle.Message:
		tell.notify(event.(oodle.Message).Nick)
	}
}

func (tell *Tell) Execute(nick string, args []string) (string, error) {
	l := &Letter{
		From: nick,
		To:   args[0],
		Body: strings.Join(args[1:], " "),
		When: time.Now(),
	}
	tell.db.Create(l)
	tell.has[nick] = true
	return "okie dokie!", nil
}
