package plugins

import (
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/godwhoa/oodle/oodle"
)

var urlReg = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)

type Title struct {
	oodle.BaseInteractive
	oodle.BaseTrigger
}

func (title *Title) OnEvent(event interface{}) {
	switch event.(type) {
	case oodle.Message:
		msg := event.(oodle.Message).Msg
		urls := urlReg.FindAllString(msg, -1)
		for _, url := range urls {
			if url != "" {
				doc, err := goquery.NewDocument(url)
				if err != nil {
					return
				}
				title.IRC.Send(doc.Find("title").First().Text())
			}
		}
	}
}
