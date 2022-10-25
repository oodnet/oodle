package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/lrstanley/girc"
	u "github.com/oodnet/oodle/utils"
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
		Name     string `json:"name"`
		Owner    struct {
			Name string `json:"name"`
		} `json:"owner"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
}

type ReleaseEvent struct {
	Action  string `json:"action"`
	Release struct {
		HTMLURL string `json:"html_url"`
		TagName string `json:"tag_name"`
		Draft   bool   `json:"draft"`
	} `json:"release"`
	Repository struct {
		FullName string `json:"full_name"`
		Name     string `json:"name"`
		Owner    struct {
			Name string `json:"name"`
		} `json:"owner"`
	} `json:"repository"`
}

// gets branch name from ref
func getBranch(ref string) string {
	fragments := strings.Split(ref, "/")
	return fragments[len(fragments)-1]
}

// shortens github urls
func gitio(ghurl string) string {
	resp, err := http.PostForm("https://git.io/create", url.Values{"url": {ghurl}})
	if err != nil {
		return ghurl
	}
	defer resp.Body.Close()
	key, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ghurl
	}
	return "https://git.io/" + string(key)
}

func pluralize(i int) string {
	if i == 1 {
		return "commit"
	}
	return "commits"
}

func (wh *webhook) handleReleaseEvent(event *ReleaseEvent) {
	repository := event.Repository.FullName
	tag := event.Release.TagName
	isDraft := event.Release.Draft
	url := gitio(event.Release.HTMLURL)
	// skip over events from repos we aren't watching
	// or it's a draft
	if _, ok := wh.watch[repository]; !ok || isDraft {
		return
	}
	msg := fmt.Sprintf("{gold}[%s]{r} %s released!: %s", repository, tag, url)
	wh.irc.Sendf(girc.Fmt(msg))
}

func (wh *webhook) handlePushEvent(event *PushEvent) {
	repository := event.Repository.FullName
	pusher := event.Pusher.Name
	compareurl := gitio(event.Compare)
	branch := getBranch(event.Ref)
	ncommits := len(event.Commits)

	// skip over events from repos/branches we aren't watching
	watching, ok := wh.watch[repository]
	if !ok || !u.Contains(watching, branch) {
		return
	}

	// <oodle> godwhoa pushed 6 new commits to master: https://git.io/fAvUI
	msg := fmt.Sprintf("{gold}[%s]{r} {green}%s{c} pushed {b}%d{r} new %s to {green}%s{c}: {blue}%s{c}", repository, pusher, ncommits, pluralize(ncommits), branch, compareurl)
	for i, commit := range event.Commits {
		if !commit.Distinct {
			continue
		}
		// only first 3 commits
		if i == 3 {
			break
		}
		// <oodle> 0f7d2e5 Godwhoa: Seperate irc client from bot package
		msg += fmt.Sprintf("\n{green}%s{c} %s: %s", commit.ID[:7], commit.Committer.Username, commit.Message)
	}
	wh.irc.Sendf(girc.Fmt(msg))
}

func (wh *webhook) Github(w http.ResponseWriter, r *http.Request) {
	hook, err := githubhook.Parse([]byte(wh.secret), r)
	if err != nil {
		http.Error(w, "Invalid secret", 400)
		return
	}

	switch hook.Event {
	case "push":
		event := &PushEvent{}
		if err := json.Unmarshal(hook.Payload, event); err != nil {
			http.Error(w, "Malformed JSON", 400)
			return
		}
		wh.handlePushEvent(event)
	case "release":
		event := &ReleaseEvent{}
		if err := json.Unmarshal(hook.Payload, event); err != nil {
			http.Error(w, "Malformed JSON", 400)
			return
		}
		wh.handleReleaseEvent(event)
	}
	w.Write([]byte(`OK`))
}
