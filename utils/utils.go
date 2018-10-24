package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/hako/durafmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/neurosnap/sentences.v1/english"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

var tokenizer, _ = english.NewSentenceTokenizer(nil)

func Summarize(text string) (summarized string) {
	sentences := tokenizer.Tokenize(text)
	for _, s := range sentences {
		if len(summarized) > 350 {
			return
		}
		summarized += s.Text
	}
	return
}

func FmtTime(t time.Time) string {
	// gets rid of milliseconds, I think?
	since := time.Since(t).Truncate(time.Second)
	// formats it to 1 day etc.
	return durafmt.Parse(since).String()
}

func FmtDur(d time.Duration) string {
	// gets rid of milliseconds, I think?
	d = d.Truncate(time.Second)
	// formats it to 1 second etc.
	return durafmt.Parse(d).String()
}

// Alias returns a new command with replaced prefix, name and usage
func Alias(prefix, name string, cmd oodle.Command) oodle.Command {
	firstFragment := strings.Split(cmd.Usage, " ")[0] // usually it's command's prefix+name
	usage := strings.Replace(cmd.Usage, firstFragment, prefix+name, 1)
	cmd.Prefix = prefix
	cmd.Name = name
	cmd.Usage = usage
	return cmd
}

// TestCase represents a test case for an oodle.Command
type TestCase struct {
	Nick  string
	Args  []string
	Reply string
	Err   error
}

// RunCases runs given tests on an oodle.Command
func RunCases(t *testing.T, cmd oodle.Command, tests []TestCase) {
	for _, tt := range tests {
		reply, err := cmd.Fn(tt.Nick, tt.Args)
		assert.Equal(t, reply, tt.Reply)
		assert.Equal(t, err, tt.Err)
	}
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
