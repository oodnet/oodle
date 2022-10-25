package urban

import (
	"strings"

	m "github.com/oodnet/oodle/middleware"
	"github.com/oodnet/oodle/oodle"
)

func Register(deps *oodle.Deps) error {
	deps.Bot.RegisterCommands(Urban())
	return nil
}

func Urban() oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "urban",
		Description: "Fetches definitions from urbandictionary.com",
		Usage:       ".urban <word>",
		Fn: func(nick string, args []string) (string, error) {
			word := strings.Join(args, " ")
			def, err := define(word)
			if err != nil {
				return "No definitions found.", nil
			}
			return def, nil
		},
	}
	cmd = m.Chain(cmd, m.MinArg(1))
	return cmd
}
