package oodle

// BaseTrigger is just an implementation of Trigger to reduce some boilerplate code
type BaseTrigger struct {
	SendQueue chan string
}

func (btrigger *BaseTrigger) SetSendQueue(sendqueue chan string) {
	btrigger.SendQueue = sendqueue
}
