package webhook

import (
	"encoding/json"
	"html"
	"net/http"

	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Register(deps *oodle.Deps) error {
	addr := viper.GetString("webhook_addr")
	if addr == "" {
		return nil
	}

	hook := &webhook{
		irc:    deps.IRC,
		log:    deps.Logger,
		secret: viper.GetString("secret"),
		watch:  viper.GetStringMapStringSlice("github"),
	}
	go hook.Listen(addr)
	return nil
}

type webhook struct {
	irc    oodle.Sender
	log    *logrus.Logger
	secret string
	// godwhoa/oodle => ["master", "dev", ...]
	watch map[string][]string
}

func (wh *webhook) Send(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("X-SECRET") != wh.secret {
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
		wh.irc.Send(girc.Fmt(html.UnescapeString(msg)))
	}
	w.Write([]byte(`OK`))
}

func (wh *webhook) Listen(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`oodle!`))
	})
	http.HandleFunc("/send", wh.Send)
	http.HandleFunc("/github", wh.Github)
	wh.log.Infof("Starting webhook on %s", addr)
	wh.log.Fatalf("Webhook failed: %v", http.ListenAndServe(addr, nil))
}
