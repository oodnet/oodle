package oodle

// BaseTrigger is just an implementation of Trigger to reduce some boilerplate code
type BaseTrigger struct {
	IRC Sender
}

func (btrigger *BaseTrigger) SetSender(sender Sender) {
	btrigger.IRC = sender
}
