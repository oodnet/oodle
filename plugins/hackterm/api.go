package hackterm

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	errNoDefinitions = errors.New("No definitions found.")
	errNoResults     = errors.New("No search results.")
)

type definitions struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	Body   []struct {
		Term string `json:"term"`
		Body string `json:"body"`
	} `json:"body"`
}

type searchResults struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	Body   []struct {
		Name string `json:"name"`
		Link string `json:"link"`
	} `json:"body"`
}

// POST makes a POST req. while pretending to be a real browser :)
func POST(url, form string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(form))
	if err != nil {
		return []byte{}, err
	}

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func search(term string) (*searchResults, error) {
	form := `term=` + term
	body, _ := POST("https://www.hackterms.com/search", form)

	results := &searchResults{}
	if err := json.Unmarshal(body, results); err != nil {
		return nil, err
	}

	if results.Status != "success" || results.Count < 1 {
		return nil, errNoResults
	}
	return results, nil
}

func getDefinitions(term string) (*definitions, error) {
	form := `term=` + term + `&user=false`
	body, _ := POST("https://www.hackterms.com/get-definitions", form)

	defs := &definitions{}
	if err := json.Unmarshal(body, defs); err != nil {
		return nil, err
	}
	if defs.Status != "success" || defs.Count < 1 {
		return nil, errNoDefinitions
	}
	return defs, nil
}

// define launches two concurrent reqs.
// req 1: search for the term, then get the definition (It has to make two req. so slower)
// req 2: directly get the definition (Just one direct req.)
// neither will send to the defCh channel unless they get it right
// we have a timeout if neither finishes
func define(term string) string {
	defCh := make(chan *definitions, 2)
	// Search then get the definition
	go func() {
		results, err := searchTerm(term)
		if err != nil {
			return
		}
		correctTerm := results.Body[0].Name
		defs, err := getDefinitions(correctTerm)
		if err != nil {
			return
		}
		defCh <- defs
	}()
	// Directly get the definition
	go func() {
		defs, err := getDefinitions(term)
		if err != nil {
			return
		}
		defCh <- defs
	}()
	// Concurrently wait on multiple channels
	select {
	case def := <-defCh:
		return def.Body[0].Term + ": " + def.Body[0].Body
	case <-time.After(2000 * time.Millisecond):
		return "No definitions found."
	}
}
