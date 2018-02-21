package main

import (
	"encoding/json"
	"html"
	"net/http"

	"github.com/godwhoa/oodle/bot"
	"github.com/lrstanley/girc"
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

func (webhook *WebHook) Send(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("X-SECRET") != webhook.secret {
		http.Error(w, "Invalid secret", 400)
		return
	}
	msgs := []string{}
	err := json.NewDecoder(r.Body).Decode(&msgs)
	if err != nil {
		http.Error(w, "Malformed JSON", 400)
		return
	}
	for _, msg := range msgs {
		webhook.irc.Send(girc.Fmt(html.UnescapeString(msg)))
	}
	w.Write([]byte(`OK`))
}

func (webhook *WebHook) Listen(addr string) error {
	http.HandleFunc("/send", webhook.Send)
	webhook.log.Infof("Starting webhooks on %s", addr)
	return http.ListenAndServe(addr, nil)
}
