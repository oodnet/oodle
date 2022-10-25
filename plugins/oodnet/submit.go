package oodnet

import (
	"net/url"
	"strings"

	m "github.com/oodnet/oodle/middleware"
	"github.com/oodnet/oodle/oodle"
	u "github.com/oodnet/oodle/utils"
	"github.com/spf13/viper"
)

func Register(deps *oodle.Deps) error {
	checker, bot := deps.IRC, deps.Bot
	key := viper.GetString("submit_key")
	if key == "" {
		return nil
	}
	bot.RegisterCommands(Submit(checker, key), Screenshot(checker, key))
	return nil
}

func Submit(checker oodle.Checker, key string) oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "submit",
		Description: "IDK, ask exezin.",
		Usage:       ".submit <desc> <url>",
		Fn: func(nick string, args []string) (string, error) {
			desc := strings.Join(args[:len(args)-1], " ")
			rawurl := args[len(args)-1]

			if _, err := url.ParseRequestURI(rawurl); err != nil {
				return "Last arg. isn't an url", nil
			}
			form := url.Values{}
			form.Set("url", rawurl)
			form.Set("username", nick)
			form.Set("text", desc)
			form.Set("password", key)
			_, err := u.HTTPClient.PostForm("https://oods.net/submit.php", form)
			if err != nil {
				return "Failed to submit: " + err.Error(), nil
			}
			return "Submitted!", nil
		},
	}
	return m.Chain(cmd, m.MinArg(2), m.RegisteredOnly(checker))
}

func Screenshot(checker oodle.Checker, key string) oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "screenshot",
		Description: "IDK, ask exezin.",
		Usage:       ".screenshot <desc> <url>",
		Fn: func(nick string, args []string) (string, error) {
			desc := strings.Join(args[:len(args)-1], " ")
			rawurl := args[len(args)-1]

			if _, err := url.ParseRequestURI(rawurl); err != nil {
				return "Last arg. isn't an url", nil
			}
			form := url.Values{}
			form.Set("url", rawurl)
			form.Set("username", nick)
			form.Set("text", desc)
			form.Set("password", key)
			_, err := u.HTTPClient.PostForm("https://oods.net/screenshot-submit", form)
			if err != nil {
				return "Failed to submit: " + err.Error(), nil
			}
			return "Submitted!", nil
		},
	}
	return m.Chain(cmd, m.MinArg(2), m.RegisteredOnly(checker))
}
