package plugins

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/godwhoa/oodle/store"
	"github.com/jmoiron/sqlx"
	"github.com/lrstanley/girc"
)

var (
	errNotInt       = errors.New("Point is not an integer.")
	errInvalidPoint = errors.New("Invalid point")
)

// Checks point given is valid
func checkPoint(str string, m map[int]time.Duration) error {
	i, err := strconv.Atoi(str)
	if err != nil {
		return errNotInt
	}
	if _, ok := m[i]; !ok {
		return errInvalidPoint
	}
	return nil
}

// Checks if a giver isn't in a timeout
func canGive(s *store.RepStore, giver string) bool {
	giverRep, err := s.GetUserRep(giver)
	return err == nil && time.Now().After(giverRep.Next)
}

type Give struct {
	store          *store.RepStore
	registeredOnly bool
	cooldowns      map[int]time.Duration
	oodle.BaseInteractive
}

func (g *Give) Configure(config *oodle.Config) {
	g.registeredOnly = config.RegisteredOnly
	g.cooldowns = make(map[int]time.Duration)
	for i, point := range config.Points {
		g.cooldowns[point] = config.Cooldowns[i].Duration
	}

}

func (g *Give) SetDB(db *sql.DB) {
	g.store = store.NewRepStore(sqlx.NewDb(db, "sqlite3"), g.cooldowns)
}

func (g *Give) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "give",
		Description: "Lets users to give rep. points to other users",
		Usage:       ".give <point> <other user>",
	}
}

func (g *Give) Execute(nick string, args []string) (string, error) {
	if len(args) != 2 || !girc.IsValidNick(args[1]) {
		return "", oodle.ErrUsage
	}
	if err := checkPoint(args[0], g.cooldowns); err != nil {
		return err.Error(), nil
	}

	giver, reciver := nick, args[1]
	point, _ := strconv.Atoi(args[0])
	if !g.IRC.InChannel(reciver) {
		return reciver + " not in channel.", nil
	}
	if g.registeredOnly && !g.IRC.IsRegistered(giver) {
		return "Only registered nicks can give.", nil
	}
	if giver == reciver {
		return "You can't give yourself points.", nil
	}
	if !canGive(g.store, giver) {
		return "You can't give points just yet.", nil
	}
	g.store.Inc(giver, reciver, point)
	reciverRep, _ := g.store.GetUserRep(reciver)
	return fmt.Sprintf("%s now has %d points!", reciverRep.User, reciverRep.Points), nil
}
