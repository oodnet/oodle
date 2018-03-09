package oodle

// BaseTrigger is just an implementation of Trigger to reduce some boilerplate code
type BaseTrigger struct{}

func (bt *BaseTrigger) OnEvent(msg interface{}) {
}

type BaseInteractive struct {
	IRC IRCClient
}

func (bi *BaseInteractive) SetIRC(irc IRCClient) {
	bi.IRC = irc
}
