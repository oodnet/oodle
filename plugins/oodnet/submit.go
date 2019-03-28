package oodnet

import (
	"net/url"
	"strings"

	m "github.com/godwhoa/oodle/middleware"
	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
)

func Submit(checker oodle.Checker) oodle.Command {
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
			from.Set("username", nick)
			from.Set("text", desc)
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
