package lazy

import (
	"net/url"
	"strings"

	"github.com/spf13/viper"

	m "github.com/godwhoa/oodle/middleware"
	"github.com/godwhoa/oodle/oodle"
)

func Register(deps *oodle.Deps) error {
	bot := deps.Bot
	irc := deps.IRC
	bot.RegisterCommands(YoutubeSearch())
	if viper.InConfig("reddit_search") {
		bot.RegisterTriggers(RedditSearch(irc))
	}
	return nil
}

func YoutubeSearch() oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "youtube",
		Description: "Gives you a search url. That's it.",
		Usage:       ".youtube <query>",
		Fn: func(nick string, args []string) (reply string, err error) {
			word := strings.Join(args, " ")
			return "https://www.youtube.com/results?search_query=" + url.PathEscape(word), nil
		},
	}
	return m.Chain(cmd, m.MinArg(1))
}
