package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/rjz/githubhook.v0"
)

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

// gets branch name from ref
func branch(ref string) string {
	fragments := strings.Split(ref, "/")
	return fragments[len(fragments)-1]
}

// shortens github urls
func gitio(ghurl string) string {
	resp, err := http.PostForm("https://git.io/create", url.Values{"url": {ghurl}})
	if err != nil {
		fmt.Println(err)
		return ghurl
	}
	defer resp.Body.Close()
	key, _ := ioutil.ReadAll(resp.Body)

	return "https://git.io/" + string(key)
}

func (wh *webhook) PushEvent(w http.ResponseWriter, r *http.Request) {
	hook, err := githubhook.Parse([]byte(wh.secret), r)
	if err != nil {
		http.Error(w, "Invalid secret", 400)
		return
	}
	payload := PushEvent{}
	if err := json.Unmarshal(hook.Payload, &payload); err != nil {
		http.Error(w, "Malformed JSON", 400)
		return
	}
	// <oodle> godwhoa pushed 6 new commits to master: https://git.io/fAvUI
	msg := fmt.Sprintf("%s pushed %d new commits to %s: %s", payload.Pusher.Name, len(payload.Commits), branch(payload.Ref), payload.Compare)
	for i, commit := range payload.Commits {
		if !commit.Distinct {
			continue
		}
		// only 3 commits
		if i == 3 {
			break
		}
		// <oodle> oodle/master 0f7d2e5 Godwhoa: Seperate irc client from bot package
		msg += fmt.Sprintf("\n%s %s %s: %s", payload.Repository.FullName, commit.ID[:6], commit.Committer.Username, commit.Message)
	}
	wh.irc.Sendf(msg)
	w.Write([]byte(`OK`))
}
