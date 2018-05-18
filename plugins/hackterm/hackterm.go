package hackterm

import (
	"strings"

	"github.com/godwhoa/oodle/oodle"
)

func Register(deps *oodle.Deps) {
	deps.Bot.Register(HackTerm())
}

func HackTerm() oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "hackterm",
		Description: "Hacker Terms",
		Usage:       ".hackterm <term>",
		Fn: func(nick string, args []string) (string, error) {
			if len(args) < 1 {
				return "", oodle.ErrUsage
			}
			term := strings.Join(args, " ")
			return define(term), nil
		},
	}
}
