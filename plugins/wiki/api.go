package wiki

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const endpoint = "https://en.wikipedia.org/w/api.php?action=opensearch&format=json&search="

var ErrNoResults = errors.New("No results.")

func search(term string) ([][]string, error) {
	response, err := http.Get(endpoint + url.QueryEscape(term))
	if err != nil {
		return [][]string{}, err
	}

	jsondata := []json.RawMessage{}
	err = json.NewDecoder(response.Body).Decode(&jsondata)
	if err != nil {
		return [][]string{}, err
	}
	// JSON structure looks something like this:
	// [string, []string, []string, []string]
	// If we remove the useless first element it'll be [][]string
	jsondata = jsondata[1:]

	searchresults := [][]string{}
	for i, jsonarray := range jsondata {
		searchresults = append(searchresults, []string{})
		json.Unmarshal(jsonarray, &searchresults[i])
	}

	// If no results are found wiki api returns 3 empty []string
	if len(searchresults) != 3 || len(searchresults[0]) < 1 {
		return [][]string{}, ErrNoResults
	}

	return searchresults, nil
}

type Extract struct {
	Term    string
	Extract string
	Link    string
}

func extract(term string) (e Extract, err error) {
	searchresults, err := search(term)
	if err != nil {
		return
	}

	terms, extracts, links := searchresults[0], searchresults[1], searchresults[2]

	// handle case where the term is ambiguous
	// see:
	// https://en.wikipedia.org/w/api.php?action=opensearch&format=json&search=Python
	i := 0
	if strings.Contains(extracts[0], "may refer to:") {
		i++
	}

	e.Term, e.Extract, e.Link = terms[i], extracts[i], links[i]

	return
}
