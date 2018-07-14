package reputation

import (
	"database/sql"
	"testing"
	"time"

	"github.com/godwhoa/oodle/oodle"

	_ "github.com/mattn/go-sqlite3"
)

func createFn(v bool) func(string) bool {
	return func(nick string) bool {
		return v
	}
}

func inmemRepStore() *RepStore {
	db, _ := sql.Open("sqlite3", ":memory:")
	return NewRepStore(db)
}

func TestValidation(t *testing.T) {
	// setup
	base := func() *Give {
		cooldowns := map[int]time.Duration{
			1:  time.Second * 1,
			-1: time.Second * 2,
			2:  time.Second * 4,
		}
		return &Give{
			inChannel:      createFn(true),
			isRegistered:   createFn(true),
			store:          inmemRepStore(),
			registeredOnly: true,
			cooldowns:      cooldowns,
		}
	}

	type testcase struct {
		title string
		nick  string
		args  []string
		reply string
		err   error
		setup func() *Give
	}
	testcases := []testcase{
		{
			title: "Can only give points to others",
			nick:  "pacninja",
			args:  []string{"1", "pacninja"},
			reply: "You can't give yourself points.",
			err:   nil,
			setup: base,
		},
		{
			title: "Bad usage",
			nick:  "pacninja",
			args:  []string{"1"},
			reply: "",
			err:   oodle.ErrUsage,
			setup: base,
		},
		{
			title: "Make you wait the timeout",
			nick:  "pacninja",
			args:  []string{"1", "apple"},
			reply: "",
			err:   oodle.ErrUsage,
			setup: func() *Give {
				g := base()
				g.cooldowns[1] = 10 * time.Second
				g.give("pacninja", []string{"1", "apple"})
				return g
			},
		},
	}
	for _, tc := range testcases {
		g := tc.setup()
		reply, err := g.give(tc.nick, tc.args)
		if tc.reply != reply {
			t.Errorf("%s failed! want: %s got: %s", tc.title, tc.reply, reply)
		}
		if tc.err != err {
			t.Errorf("%s failed! want: %s got: %s", tc.title, tc.err, err)
		}
	}
}
