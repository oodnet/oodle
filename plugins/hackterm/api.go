package hackterm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	u "github.com/godwhoa/oodle/utils"
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

func search(term string) (*searchResults, error) {
	form := `term=` + term
	body, err := u.FakePOST("https://www.hackterms.com/search", strings.NewReader(form))
	if err != nil {
		return nil, err
	}

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
	body, err := u.FakePOST("https://www.hackterms.com/get-definitions", strings.NewReader(form))
	fmt.Println(string(body))
	if err != nil {
		return nil, err
	}

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
		results, err := search(term)
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
