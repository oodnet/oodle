package plugins

import (
	"fmt"
	"strings"

	"github.com/godwhoa/oodle/oodle"
)

func isSed(msg string) bool {
	return strings.HasPrefix(msg, "s/") && strings.Count(msg, "/") > 2
}

func sed(msg string, script string) string {
}

type Sed struct {
	lastMsg string
	oodle.BaseInteractive
}

func (s *Sed) OnEvent(event interface{}) {
	message, ok := event.(oodle.Message)
	if !ok {
		return
	}
	if isSed(message.Msg) {
		s.IRC.Send(sed(s.lastMsg, message.Msg))
		return
	}
	s.lastMsg = fmt.Sprintf("<%s> %s", message.Nick, message.Msg)
}
