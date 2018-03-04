package plugins

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/godwhoa/oodle/oodle"
)

var (
	errNoDefinitions = errors.New("No definitions found.")
	errNoResults     = errors.New("No search results.")
)

type Definitions struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	Body   []struct {
		Body string `json:"body"`
	} `json:"body"`
}

type SearchResults struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	Body   []struct {
		Name string `json:"name"`
		Link string `json:"link"`
	} `json:"body"`
}

// Lets pretend to be a browser :)
func setHeaders(req *http.Request) {
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Origin", "https://www.hackterms.com")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.119 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://www.hackterms.com/")
}

func searchTerm(term string) (*SearchResults, error) {
	form := `term=` + term
	req, err := http.NewRequest("POST", "https://www.hackterms.com/search", strings.NewReader(form))
	if err != nil {
		return nil, err
	}
	setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	results := &SearchResults{}
	if err := json.Unmarshal(body, results); err != nil {
		return nil, err
	}

	if results.Status != "success" || results.Count < 1 {
		return nil, errNoResults
	}
	return results, nil
}

func getDefinition(term string) (*Definitions, error) {
	form := `term=` + term + `&user=false`
	req, err := http.NewRequest("POST", "https://www.hackterms.com/get-definitions", strings.NewReader(form))
	if err != nil {
		return nil, err
	}
	setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	defs := &Definitions{}
	if err := json.Unmarshal(body, defs); err != nil {
		return nil, err
	}
	if defs.Status != "success" || defs.Count < 1 {
		return nil, errNoDefinitions
	}
	return defs, nil
}

// findDefinition launches two concurrent reqs.
// req 1: search for the term, then get the definition (It has two make two req. so slower)
// req 2: directly get the definition (Just one direct req.)
// neither will send to the defCh channel unless they get it right
// we have a timeout if neither finishes
func findDefinition(term string) string {
	defCh := make(chan *Definitions, 2)
	// Search then get the definition
	go func() {
		results, err := searchTerm(term)
		if err != nil {
			return
		}
		correctTerm := results.Body[0].Name
		defs, err := getDefinition(correctTerm)
		if err != nil {
			return
		}
		defCh <- defs
	}()
	// Directly get the definition
	go func() {
		defs, err := getDefinition(term)
		if err != nil {
			return
		}
		defCh <- defs
	}()
	// concurrently wait on multiple channels
	select {
	case def := <-defCh:
		return def.Body[0].Body
	case <-time.After(1500 * time.Millisecond):
		return "No definitions found."
	}
}

type HackTerm struct {
	oodle.BaseTrigger
}

func (hackterm *HackTerm) Info() oodle.CommandInfo {
	return oodle.CommandInfo{
		Prefix:      ".",
		Name:        "hackterm",
		Description: "Hacker Terms",
		Usage:       ".hackterm <term>",
	}
}

func (hackterm *HackTerm) Execute(nick string, args []string) (string, error) {
	if len(args) < 1 {
		return "", oodle.ErrUsage
	}
	term := strings.Join(args, " ")
	return findDefinition(term), nil
}
