package lazy

import (
	"net/url"
	"regexp"

	"fmt"

	"strings"

	"github.com/lrstanley/girc"
	"github.com/oodnet/oodle/events"
	"github.com/oodnet/oodle/oodle"
	u "github.com/oodnet/oodle/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Pattern struct {
	regex     *regexp.Regexp
	subreddit string
	query     string
}

type SearchResult struct {
	Title     string `json:"title"`
	Permalink string `json:"permalink"`
	URL       string `json:"url"`
}

type SearchContainer struct {
	Kind string `json:"kind"`
	Data struct {
		Children []struct {
			Data SearchResult `json:"data,omitempty"`
		} `json:"children"`
	} `json:"data"`
}

func search(subreddit, query string) (results []SearchResult, err error) {
	query = url.QueryEscape(query)
	url := fmt.Sprintf(`https://www.reddit.com/%s/search.json?q=%s&restrict_sr=true`, subreddit, query)
	container := SearchContainer{}
	err = u.GetJSON(url, &container)
	results = []SearchResult{}
	for _, child := range container.Data.Children {
		results = append(results, child.Data)
	}
	return
}

func RedditSearch(sender oodle.Sender) oodle.Trigger {
	config := viper.GetStringMapStringSlice("reddit_search")
	patterns := []Pattern{}
	for subreddit, searchtuple := range config {
		if !strings.HasPrefix(subreddit, "r/") {
			logrus.Errorf("Invalid subreddit: %s", subreddit)
			continue
		}
		if len(searchtuple) < 2 {
			logrus.Errorf("Incorrect configuration for %s in reddit_search", subreddit)
			continue
		}
		regex, err := regexp.Compile(searchtuple[0])
		if err != nil {
			logrus.Errorf("Regex did not compile for: %s  %s", subreddit, searchtuple[1])
			continue
		}
		patterns = append(patterns, Pattern{
			regex:     regex,
			subreddit: subreddit,
			query:     searchtuple[1],
		})
	}

	trigger := func(event girc.Event) {
		if !events.Is(event, events.MESSAGE) {
			return
		}
		_, message := events.Message(event)
		for _, p := range patterns {
			fmt.Println(message, p.regex.MatchString(message))
			if !p.regex.MatchString(message) {
				continue
			}
			results, err := search(p.subreddit, p.query)
			if err != nil {
				logrus.Error(err)
				return
			}
			if len(results) < 1 {
				return
			}
			sender.Sendf("Title: %s URL: %s", results[0].Title, results[1].URL)
		}

	}
	return oodle.Trigger(trigger)
}
