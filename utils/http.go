package utils

import (
	"net"
	"net/http"
	"time"
)

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

var HTTPClient = &http.Client{
	Transport: netTransport,
	Timeout:   2 * time.Second,
}
