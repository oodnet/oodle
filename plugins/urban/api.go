package urban

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type definition struct {
	Word       string `json:"word"`
	Permalink  string `json:"permalink"`
	ThumbsUp   int    `json:"thumbs_up"`
	Definition string `json:"definition"`
	ThumbsDown int    `json:"thumbs_down"`
}

type results struct {
	List []definition `json:"list"`
}

func top(defs []definition) definition {
	topscore := 0
	topdef := definition{}
	for _, d := range defs {
		score := d.ThumbsUp - d.ThumbsDown
		if score > topscore {
			topscore = score
			topdef = d
		}
	}
	return topdef
}

func format(def definition) string {
	lines := strings.Split(def.Definition, "\r\n\r\n")
	if len(lines) > 3 {
		lines = lines[:3]
	}

	trimmed := strings.Join(lines, " ")
	return fmt.Sprintf("%s: %s\n%s", def.Word, def.Permalink, trimmed)
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

	return format(r.List[0]), nil
}
