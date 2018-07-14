package reputation

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
	"github.com/lrstanley/girc"
)

func pluralize(word string, i int) string {
	if i == 1 {
		return word
	}
	return word + "s"
}

func parse(config *oodle.Config) map[int]time.Duration {
	cooldowns := make(map[int]time.Duration)
	for i, point := range config.Points {
		cooldowns[point] = config.Cooldowns[i].Duration
	}
	return cooldowns
}

func Register(deps *oodle.Deps) error {
	if len(deps.Config.Cooldowns) != len(deps.Config.Points) {
		return errors.New("len(cooldowns) != len(points)")
	}

	cooldowns := parse(deps.Config)
	store := NewRepStore(deps.DB)
	g := &Give{
		inChannel:      deps.IRC.InChannel,
		isRegistered:   deps.IRC.IsRegistered,
		store:          store,
		registeredOnly: deps.Config.RegisteredOnly,
		cooldowns:      cooldowns,
	}
	deps.Bot.Register(
		Rank(store),
		oodle.Command{
			Prefix:      ".",
			Name:        "give",
			Description: "Lets users to give rep. points to other users",
			Usage:       ".give <point> <other user>",
			Fn:          g.give,
		},
		oodle.Command{
			Prefix:      ".",
			Name:        "rep",
			Description: "Lets you give 1 rep. point to other user",
			Usage:       ".rep <other user>",
			Fn:          g.rep,
		},
	)
	return nil
}

func Rank(store *RepStore) oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "rank",
		Description: "Lets you check rep. points of a user",
		Usage:       ".rank <user>",
		Fn: func(nick string, args []string) (string, error) {
			if len(args) != 1 {
				return "", oodle.ErrUsage
			}
			user := args[0]
			points := store.Points(user)
			return fmt.Sprintf("%s has %d %s.", user, points, pluralize("point", points)), nil
		},
	}
}

type Give struct {
	inChannel      func(nick string) bool
	isRegistered   func(nick string) bool
	registeredOnly bool
	store          *RepStore
	cooldowns      map[int]time.Duration
}

func (g *Give) canGive(giver string) (bool, time.Duration) {
	point, lastGiven := g.store.LastGiven(giver)
	timeout := g.cooldowns[point]
	// fmt.Printf("\npoint: %d timeout: %s\n", point, timeout.String())
	can := time.Now().After(lastGiven.Add(timeout))
	wait := lastGiven.Add(timeout).Sub(time.Now())
	// fmt.Printf("lastgiven + timeout = %s\n", u.FmtTime(lastGiven))
	// fmt.Printf("wait: %s can: %v", u.FmtDur(wait), can)
	return can, wait
}

// validate giver/reciver
func (g *Give) validateGR(giver, reciver string) string {
	if giver == reciver {
		return "You can't give yourself points."
	}
	if can, wait := g.canGive(giver); !can {
		return "You can't give points just yet. Wait " + u.FmtDur(wait)
	}
	if !g.inChannel(reciver) {
		return reciver + " is not in the channel."
	}
	if g.registeredOnly && !g.isRegistered(giver) {
		return "Only registered nicks can give."
	}
	return ""
}

func (g *Give) give(nick string, args []string) (string, error) {
	// Command validation
	if len(args) != 2 || !girc.IsValidNick(args[1]) {
		return "", oodle.ErrUsage
	}

	// Point/Cooldown validation
	point, err := strconv.Atoi(args[0])
	if err != nil {
		return "First argument need to be an integer.", nil
	}
	if _, ok := g.cooldowns[point]; !ok {
		return "Invalid point", nil
	}

	giver, reciver := nick, args[1]
	if errmsg := g.validateGR(giver, reciver); errmsg != "" {
		return errmsg, nil
	}

	reciverPoints, _ := g.store.Give(giver, reciver, point)
	return fmt.Sprintf("%s now has %d points!", reciver, reciverPoints), nil
}

func (g *Give) rep(nick string, args []string) (string, error) {
	if len(args) != 1 {
		return "", oodle.ErrUsage
	}

	if _, ok := g.cooldowns[1]; !ok {
		return "Invalid point", nil
	}

	giver, reciver := nick, args[0]
	if errmsg := g.validateGR(giver, reciver); errmsg != "" {
		return errmsg, nil
	}

	reciverPoints, _ := g.store.Give(nick, args[0], 1)
	return fmt.Sprintf("%s now has %d %s!", args[0], reciverPoints, pluralize("point", reciverPoints)), nil
}
