package sed

import (
	"strings"

	"github.com/godwhoa/oodle/oodle"
)

func Register(deps *oodle.Deps) error {
	bot, irc := deps.Bot, deps.IRC
	bot.RegisterTriggers(Sed(irc))
	bot.RegisterCommands(SedHelp())
	return nil
}

func SedHelp() oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "sed",
		Description: "Simple sed-like find and replace. Since it is simple you can't replace '/'",
		Usage:       "s/{find}/{replace}/ ending '/' is optional",
		Fn: func(nick string, args []string) (reply string, err error) {
			return ".sed is not an actual command. Do s/{find}/{replace} to use it. This exists since murii might do .help sed", nil
		},
	}
}

func Sed(sender oodle.Sender) oodle.Trigger {
	usermsgs := make(map[string]string)
	return func(event interface{}) {
		message, ok := event.(oodle.Message)
		if !ok {
			return
		}
		msg, nick := message.Msg, message.Nick
		// Check if it is a sed command
		if strings.HasPrefix(msg, "s/") && strings.Count(msg, "/") >= 2 {
			args := strings.Split(msg, "/")
			// {0}/{1}/{2}
			if len(args) < 3 {
				return
			}
			// apply find n replace on user's last msg.
			lastmsg, ok := usermsgs[nick]
			if !ok {
				sender.Sendf("Your last message was not found.")
				return
			}
			newmsg := strings.Replace(lastmsg, args[1], args[2], -1)
			sender.Sendf("<%s> %s", nick, newmsg)
		} else {
			// store it as user's last msg.
			usermsgs[nick] = msg
		}
	}
}
