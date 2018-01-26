package oodle

// Simpified IRC events; they'll be sent thru eventboxes
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
