package plugins

import (
	"strings"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/hako/durafmt"
	"github.com/jinzhu/gorm"
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
	db *gorm.DB
	// cache for checking if a user has mail
	// to avoid querying for every msg.
	lcount map[string]int
	oodle.BaseTrigger
}

func (tell *Tell) notify(nick string) {
	if c, ok := tell.lcount[nick]; !ok || c < 1 {
		return
	}
	var letters []Letter
	tell.db.Where("`to` = ?", nick).Find(&letters)
	for _, l := range letters {
		tell.IRC.Sendf("%s, %s left this message for you: %s\n%s ago", nick, l.From, l.Body, fmtTime(l.When))
		tell.db.Delete(&l)
		tell.lcount[nick]--
	}
}

func (tell *Tell) Init(config *oodle.Config, db *gorm.DB) {
	db.AutoMigrate(&Letter{})
	tell.db = db
	tell.lcount = make(map[string]int)

	// update cache
	var letters []Letter
	db.Select("`to`").Find(&letters)
	for _, l := range letters {
		tell.lcount[l.To]++
	}
}

func (tell *Tell) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "tell",
		Description: "Lets you send a msg. to an inactive user.\nIt will notify them once they are active.",
		Usage:       ".tell <to> <msg>",
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
	if len(args) < 2 {
		return "", oodle.ErrUsage
	}
	l := &Letter{
		From: nick,
		To:   args[0],
		Body: strings.Join(args[1:], " "),
		When: time.Now(),
	}
	tell.db.Create(l)
	tell.lcount[args[0]]++
	return "okie dokie!", nil
}
