package plugins

import (
	"fmt"

	"github.com/godwhoa/oodle/oodle"
	"github.com/jinzhu/gorm"
	"github.com/lrstanley/girc"
)

type Rank struct {
	db *gorm.DB
}

func (rank *Rank) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "rank",
		Description: "Lets you check rep. points of a user",
		Usage:       ".rank <user>",
	}
}

func (rank *Rank) Init(config *oodle.Config, db *gorm.DB) {
	rank.db = db
}

func (rank *Rank) Execute(nick string, args []string) (string, error) {
	if len(args) < 1 || !girc.IsValidNick(args[0]) {
		return "", oodle.ErrUsage
	}
	db := rank.db
	userRep := &Reputation{}
	if db.Where(Reputation{User: args[0]}).Find(userRep).RecordNotFound() {
		return fmt.Sprintf("No records for %s in db.", args[0]), nil
	}
	return fmt.Sprintf("%s has %d points", userRep.User, userRep.Points), nil
}
