package oodle

// BaseTrigger is just an implementation of Trigger to reduce some boilerplate code
type BaseTrigger struct {
	IRC IRCClient
}

func (btrigger *BaseTrigger) OnEvent(msg interface{}) {
}

func (btrigger *BaseTrigger) SetIRC(irc IRCClient) {
	btrigger.IRC = irc
}
