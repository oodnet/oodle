package plugins

import (
	"fmt"

	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
)

type Rep struct {
	Give
}

func (r *Rep) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "rep",
		Description: "Lets you give 1 rep. point to other user",
		Usage:       ".rep <other user>",
	}
}

func (r *Rep) Execute(nick string, args []string) (string, error) {
	if len(args) != 1 || !girc.IsValidNick(args[0]) {
		return "", oodle.ErrUsage
	}
	if _, ok := r.cooldowns[1]; !ok {
		return "cooldowns/points not configured properly.", nil
	}

	giver, reciver, point := nick, args[0], 1

	if m := r.Validate(giver, reciver); m != "" {
		return m, nil
	}

	r.store.Inc(giver, reciver, point)
	reciverRep, _ := r.store.GetUserRep(reciver)
	return fmt.Sprintf("%s now has %d points!", reciverRep.User, reciverRep.Points), nil
}
