package events

import (
	"github.com/lrstanley/girc"
)

type Kind uint

// Simplified set of event "kinds"
const (
	CONNECTED Kind = iota // connected to the irc server
	JOINED                // join the irc channel
	JOIN                  // user joins
	LEAVE                 // user leaves
	MESSAGE               // irc message
)

var mapping = map[Kind]string{
	CONNECTED: girc.RPL_WELCOME,
	JOINED:    girc.RPL_CHANNELMODEIS,
	JOIN:      girc.JOIN,
	LEAVE:     girc.PART,
	MESSAGE:   girc.PRIVMSG,
}

// Is checks if the event is of the kind we want
func Is(event girc.Event, kind Kind) bool {
	command := mapping[kind]
	return event.Command == command
}

func Any(event girc.Event, kinds ...Kind) bool {
	for _, k := range kinds {
		if Is(event, k) {
			return true
		}
	}
	return false
}

// Nick extracts nick
func Nick(event girc.Event) string {
	return event.Source.Name
}

// Message extracts nick and message
func Message(event girc.Event) (string, string) {
	return event.Source.Name, event.Trailing
}

// Join extracts nick and channel
func Join(event girc.Event) (string, string) {
	return event.Source.Name, event.Trailing
}

// Leave extracts nick and channel
func Leave(event girc.Event) (string, string) {
	return event.Source.Name, event.Trailing
}
