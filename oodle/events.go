package oodle

// Simpified IRC events
type Message struct {
	Nick string
	Msg  string
}

type Join struct {
	Nick string
}

type Leave struct {
	Nick string
}

type Joined struct{}
