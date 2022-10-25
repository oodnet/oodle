package core

import (
	"time"

	"github.com/oodnet/oodle/oodle"
	u "github.com/oodnet/oodle/utils"
)

type Release struct {
	HTMLURL     string    `json:"html_url"`
	ID          int       `json:"id"`
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		URL       string    `json:"url"`
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		State     string    `json:"state"`
		Size      int       `json:"size"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"assets"`
}

type Updates struct {
	sender   oodle.Sender
	notified map[string]bool
}

func (up *Updates) Notify() {
	release := Release{}
	err := u.GetJSON("https://api.github.com/repos/oodnet/oodle/releases/latest", &release)
	if err != nil {
		return // log at least at debug level maybe?
	}
	// skip unfinished release
	if release.Prerelease || release.Draft || len(release.Assets) < 1 {
		return
	}
	// skip if we already notified
	if _, ok := up.notified[release.TagName]; ok {
		return
	}
	// unnotifed new release!
	if oodle.Version != release.TagName {
		up.sender.Sendf("New release %s  %s!", release.TagName, release.HTMLURL)
	}
}

func (up *Updates) Loop() {
	for {
		up.Notify()
		time.Sleep(time.Hour * 24)
	}
}
