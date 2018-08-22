package urban

import (
	"encoding/json"
	"errors"
	"net/http"
)

type results struct {
	List []struct {
		ThumbsUp   int    `json:"thumbs_up"`
		Definition string `json:"definition"`
		ThumbsDown int    `json:"thumbs_down"`
	} `json:"list"`
}

func top(r results) string {
	topscore := 0
	topdef := ""
	for _, d := range r.List {
		score := d.ThumbsUp - d.ThumbsDown
		if score > topscore {
			topscore = score
			topdef = d.Definition
		}
	}
	return topdef
}

var ErrNoDefinition = errors.New("No definition found.")

func define(word string) (string, error) {
	response, err := http.Get("https://api.urbandictionary.com/v0/define?term=" + word)
	if err != nil {
		return "", err
	}

	r := results{}
	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return "", err
	}

	if len(r.List) < 1 {
		return "", ErrNoDefinition
	}

	return top(r), nil
}
