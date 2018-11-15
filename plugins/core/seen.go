package core

import (
	"fmt"
	"time"

	m "github.com/godwhoa/oodle/middleware"
	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
	"github.com/jmoiron/sqlx"
)

// Seen tells you when it last saw someone
type Seen struct {
	db *sqlx.DB
}

func (s *Seen) Migrate() {
	stmt := `
	CREATE TABLE IF NOT EXISTS "seen" (
		"id" integer primary key autoincrement,
		"nick" varchar(255) UNIQUE,
		"when" datetime
	);
	`
	s.db.MustExec(stmt)
}

func (s *Seen) update(nick string) {
	stmt := `INSERT INTO seen("nick", "when") VALUES(?,?) ON CONFLICT(nick) DO UPDATE SET "when" = ?;`
	s.db.Exec(stmt, nick, time.Now(), time.Now())
}

func (s *Seen) get(nick string) (lastseen time.Time, err error) {
	err = s.db.QueryRow(`SELECT "when" FROM seen WHERE "nick" = ?`, nick).
		Scan(&lastseen)
	return
}

func (s *Seen) Trigger() oodle.Trigger {
	return func(event interface{}) {
		nick := ""
		switch event.(type) {
		case oodle.Join:
			nick = event.(oodle.Join).Nick
		case oodle.Leave:
			nick = event.(oodle.Leave).Nick
		case oodle.Message:
			nick = event.(oodle.Message).Nick
		}
		if nick != "" {
			s.update(nick)
		}
	}
}

func (s *Seen) Command() oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "seen",
		Description: "Tells you when it last saw someone",
		Usage:       ".seen <nick>",
		Fn: func(nick string, args []string) (string, error) {
			if lastSeen, err := s.get(args[0]); err == nil {
				return fmt.Sprintf("%s was last seen %s ago.", args[0], u.FmtTime(lastSeen)), nil
			}
			return fmt.Sprintf("No logs for %s", args[0]), nil
		},
	}
	cmd = m.Chain(cmd, m.MinArg(1))
	return cmd
}
