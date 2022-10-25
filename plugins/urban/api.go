package urban

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	u "github.com/oodnet/oodle/utils"
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
	def.Definition = strings.Replace(def.Definition, "\r\n", "", -1)
	def.Definition = u.Summarize(def.Definition)
	return fmt.Sprintf("%s: %s\n%s", def.Word, def.Permalink, def.Definition)
}

var ErrNoDefinition = errors.New("No definition found.")

func define(word string) (string, error) {
	r := results{}
	err := u.GetJSON("https://api.urbandictionary.com/v0/define?term="+url.QueryEscape(word), &r)
	if err != nil {
		return "", err
	}

	if len(r.List) < 1 {
		return "", ErrNoDefinition
	}

	return format(r.List[0]), nil
}
