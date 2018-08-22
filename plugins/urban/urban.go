package urban

import (
	"strings"

	"github.com/godwhoa/oodle/oodle"
)

func Register(deps *oodle.Deps) error {
	deps.Bot.Register(Urban())
	return nil
}

func Urban() oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "urban",
		Description: "Fetches definitions from urbandictionary.com",
		Usage:       ".urban <word>",
		Fn: func(nick string, args []string) (string, error) {
			if len(args) < 1 {
				return "", oodle.ErrUsage
			}
			word := strings.Join(args, " ")
			def, err := define(word)
			if err != nil {
				return "No definitions found.", nil
			}
			return def, nil
		},
	}
}
