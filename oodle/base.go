package oodle

// BaseTrigger is just an implementation of Trigger to reduce some boilerplate code
type BaseTrigger struct {
	sendqueue chan string
}

func (btrigger *BaseTrigger) SetSendQueue(sendqueue chan string) {
	btrigger.sendqueue = sendqueue
}
