package webhook

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/viper"

	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

func Register(deps *oodle.Deps) error {
	addr := viper.GetString("webhook_addr")
	if addr == "" {
		return nil
	}
	hook := &webhook{irc: deps.IRC, log: deps.Logger, secret: viper.GetString("secret")}
	go hook.Listen(addr)
	return nil
}

type PushEvent struct {
	Ref     string `json:"ref"`
	Compare string `json:"compare"`
	Commits []struct {
		ID        string `json:"id"`
		Distinct  bool   `json:"distinct"`
		Message   string `json:"message"`
		Committer struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"committer"`
	} `json:"commits"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
}

type webhook struct {
	irc    oodle.IRCClient
	log    *logrus.Logger
	secret string
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

func branch(ref string) string {
	fragments := strings.Split(ref, "/")
	return fragments[len(fragments)-1]
}

func gitio(ghurl string) string {
	resp, err := http.PostForm("https://git.io/create", url.Values{"url": {ghurl}})
	if err != nil {
		return ghurl
	}
	defer resp.Body.Close()
	key, _ := ioutil.ReadAll(resp.Body)

	return "https://git.io/" + string(key)
}

func (wh *webhook) PushEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("X-SECRET") != wh.secret {
		http.Error(w, "Invalid secret", 400)
		return
	}
	payload := PushEvent{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Malformed JSON", 400)
		return
	}
	msg := fmt.Sprintf("%s pushed %d new commits to %s: %s", payload.Pusher.Name, len(payload.Commits), branch(payload.Ref), payload.Compare)
	for i, commit := range payload.Commits {
		if !commit.Distinct {
			continue
		}
		if i == 3 {
			break
		}
		msg += fmt.Sprintf("\n%s %s %s: %s", payload.Repository.FullName, commit.ID[:6], commit.Committer.Username, commit.Message)
	}
	wh.irc.Sendf(msg)
	w.Write([]byte(`OK`))
}

func (wh *webhook) Listen(addr string) {
	http.HandleFunc("/send", wh.Send)
	wh.log.Infof("Starting webhook on %s", addr)
	wh.log.Fatalf("Webhook failed: %v", http.ListenAndServe(addr, nil))
}
