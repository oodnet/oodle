package plugins

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/store"
	"github.com/jmoiron/sqlx"
	"github.com/lrstanley/girc"
)

type Rank struct {
	store *store.RepStore
}

func (r *Rank) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "rank",
		Description: "Lets you check rep. points of a user",
		Usage:       ".rank <user>",
	}
}

func (r *Rank) SetDB(db *sql.DB) {
	r.store = store.NewRepStore(sqlx.NewDb(db, "sqlite3"), make(map[int]time.Duration))
}

func (r *Rank) Execute(nick string, args []string) (string, error) {
	if len(args) < 1 || !girc.IsValidNick(args[0]) {
		return "", oodle.ErrUsage
	}
	userRep, _ := r.store.GetUserRep(args[0])
	return fmt.Sprintf("%s has %d points", userRep.User, userRep.Points), nil
}
