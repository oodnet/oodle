package wiki

import (
	"fmt"
	"strings"

	m "github.com/godwhoa/oodle/middleware"
	"github.com/godwhoa/oodle/oodle"
)

func Register(deps *oodle.Deps) error {
	deps.Bot.RegisterCommands(Wiki())
	return nil
}

func format(e Extract) string {
	return fmt.Sprintf("%s: %s\n%s", e.Term, e.Link, e.Extract)
}

func Wiki() oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "wiki",
		Description: "Fetches extracts from wikipedia",
		Usage:       ".wiki <word>",
		Fn: func(nick string, args []string) (string, error) {
			word := strings.Join(args, " ")
			e, err := extract(word)
			if err != nil {
				return err.Error(), nil
			}
			return format(e), nil
		},
	}
	cmd = m.Chain(cmd, m.MinArg(1))
	return cmd
}
