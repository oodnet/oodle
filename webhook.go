package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"github.com/godwhoa/oodle/bot"
	"github.com/sirupsen/logrus"
)

type WebHook struct {
	irc    *bot.IRCClient
	log    *logrus.Logger
	secret string
}

func NewWebHook(irc *bot.IRCClient, log *logrus.Logger, secret string) *WebHook {
	return &WebHook{
		irc:    irc,
		log:    log,
		secret: secret,
	}
}

type Thread struct {
	User      string `json:"user"`
	Title     string `json:"title"`
	Permalink string `json:"permalink"`
}

func (webhook *WebHook) NewThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("X-SECRET") != webhook.secret {
		http.Error(w, "Invalid secret", 400)
		return
	}
	t := &Thread{}
	err := json.NewDecoder(r.Body).Decode(t)
	if err != nil {
		http.Error(w, "Malformed JSON", 400)
		return
	}
	webhook.irc.Send(fmt.Sprintf(`oods.net: %s started thread "%s"`, t.User, html.UnescapeString(t.Title)))
	webhook.irc.Send(t.Permalink)
	w.Write([]byte(`OK`))
}

func (webhook *WebHook) Listen(addr string) error {
	http.HandleFunc("/event/thread", webhook.NewThread)
	webhook.log.Infof("Starting webhooks on %s", addr)
	return http.ListenAndServe(addr, nil)
}
