package plugins

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/jinzhu/gorm"
	"github.com/lrstanley/girc"
)

type Reputation struct {
	gorm.Model
	User   string
	Points int
	Next   time.Time
}

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
func canGive(db *gorm.DB, giver string) bool {
	giverRep := &Reputation{}
	db.Where(Reputation{User: giver}).
		Attrs(Reputation{
			Points: 0,
			Next:   time.Now().Add(-1 * time.Second),
		}).
		FirstOrCreate(giverRep)
	return time.Now().After(giverRep.Next)
}

type Give struct {
	db *gorm.DB
	oodle.BaseTrigger
	registeredOnly bool
	cooldowns      map[int]time.Duration
}

func (give *Give) Init(config *oodle.Config, db *gorm.DB) {
	db.AutoMigrate(&Reputation{})
	give.db = db
	give.registeredOnly = config.RegisteredOnly
	give.cooldowns = make(map[int]time.Duration)
	for i, point := range config.Points {
		give.cooldowns[point] = config.Cooldowns[i].Duration
	}
}

func (give *Give) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "give",
		Description: "Lets users to give rep. points to other users",
		Usage:       ".give <point> <other user>",
	}
}

func (give *Give) give(giver, reciver string, point int) {
	giverRep, reciverRep := &Reputation{}, &Reputation{}
	next := time.Now().Add(give.cooldowns[point])
	db := give.db
	// TODO: make this more efficent.
	// increment rep for reciver
	db.Where(Reputation{User: reciver}).
		Attrs(Reputation{
			Next: time.Now().Add(-1 * time.Second),
		}).
		FirstOrCreate(reciverRep)
	db.Model(reciverRep).
		UpdateColumn("points", gorm.Expr("points + ?", point))
	// add cooldown for giver
	db.Model(giverRep).Where(Reputation{User: giver}).
		Update("next", next)
}

func (give *Give) Execute(nick string, args []string) (string, error) {
	if len(args) != 2 || !girc.IsValidNick(args[1]) {
		return "", oodle.ErrUsage
	}
	if err := checkPoint(args[0], give.cooldowns); err != nil {
		return err.Error(), nil
	}
	giver, reciver := nick, args[1]
	point, _ := strconv.Atoi(args[0])
	if !give.IRC.InChannel(reciver) {
		return reciver + " not in channel.", nil
	}
	if give.registeredOnly && !give.IRC.IsRegistered(giver) {
		return "Only registered nicks can give.", nil
	}
	if giver == reciver {
		return "You can't give yourself points.", nil
	}
	if !canGive(give.db, giver) {
		return "You can't give points just yet.", nil
	}
	give.give(giver, reciver, point)
	reciverRep := &Reputation{}
	give.db.Where(Reputation{User: reciver}).Find(reciverRep)
	return fmt.Sprintf("%s now has %d points!", reciverRep.User, reciverRep.Points), nil
}
