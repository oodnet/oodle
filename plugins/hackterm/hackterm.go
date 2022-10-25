package hackterm

import (
	"strings"

	m "github.com/oodnet/oodle/middleware"
	"github.com/oodnet/oodle/oodle"
)

func Register(deps *oodle.Deps) error {
	deps.Bot.RegisterCommands(HackTerm())
	return nil
}

func HackTerm() oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "hackterm",
		Description: "Fetches definitions from hackerterms.com",
		Usage:       ".hackterm <term>",
		Fn: func(nick string, args []string) (string, error) {
			term := strings.Join(args, " ")
			return define(term), nil
		},
	}
	cmd = m.Chain(cmd, m.MinArg(1))
	return cmd
}
