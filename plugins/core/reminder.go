package core

import (
	"database/sql"
	"strings"
	"time"

	m "github.com/godwhoa/oodle/middleware"
	"github.com/godwhoa/oodle/oodle"
	u "github.com/godwhoa/oodle/utils"
	"github.com/jmoiron/sqlx"
)

/*
> Persist reminders to database
> A loop queries the database for reminders which need to be sent out every 1sec
> If user is in channel then send it out
> If not add it to .tell's store
*/

type RemindIn struct {
	checker  oodle.Checker
	sender   oodle.Sender
	store    *ReminderStore
	mailbox  *MailBox
	queryNow chan bool
}

const pollrate = 6 * time.Hour

func (r *RemindIn) send(reminder Reminder) {
	if !r.checker.InChannel(reminder.By) {
		r.mailbox.Send(Letter{
			From: "reminder_system",
			To:   reminder.By,
			Body: reminder.Msg,
			When: time.Now(),
		})
		return
	}
	r.sender.Sendf("%s, Reminder: %s", reminder.By, reminder.Msg)
}

func (r *RemindIn) fn(nick string, args []string) (string, error) {
	duration, err := u.ParseDuration(args[0])
	if err != nil {
		return err.Error(), nil
	}
	msg := strings.Join(args[1:], " ")
	reminder := Reminder{
		By:  nick,
		Msg: msg,
		At:  time.Now().Add(duration),
	}
	r.store.Set(reminder)
	time.AfterFunc(duration, func() {
		r.queryNow <- true
	})
	return "Reminder set!", nil
}

func (r *RemindIn) sendout() {
	reminders := r.store.Reminders()
	for _, reminder := range reminders {
		if time.Now().After(reminder.At) {
			r.send(reminder)
			r.store.Delete(reminder.ID)
		}
	}
}

// Watch watches the store and sends out reminder that need to be out
func (r *RemindIn) Watch() {
	t := time.NewTicker(pollrate)
	for {
		select {
		case <-r.queryNow:
			r.sendout()
		case <-t.C:
			r.sendout()
		}
	}
}

func (r *RemindIn) scheduleExisting() {
	reminders := r.store.Reminders()
	for _, reminder := range reminders {
		duration := reminder.At.Sub(time.Now())
		time.AfterFunc(duration, func() {
			r.queryNow <- true
		})
	}
}

func (r *RemindIn) Command() oodle.Command {
	cmd := oodle.Command{
		Prefix:      ".",
		Name:        "remindin",
		Description: "Lets you set yourself a reminder",
		Usage:       ".remindin <duration> <msg>",
		Fn:          r.fn,
	}
	cmd = m.Chain(cmd, m.MinArg(2))
	return cmd
}

type Reminder struct {
	ID        uint       `db:"id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
	By        string     `db:"by"`
	Msg       string     `db:"msg"`
	At        time.Time  `db:"at"`
}

type ReminderStore struct {
	db *sqlx.DB
}

func NewReminderStore(db *sql.DB) *ReminderStore {
	r := &ReminderStore{
		db: sqlx.NewDb(db, "sqlite3"),
	}
	r.Migrate()
	return r
}

func (r *ReminderStore) Migrate() {
	stmt := `
	CREATE TABLE IF NOT EXISTS "reminders" (
		"id" integer primary key autoincrement,
		"created_at" datetime,
		"updated_at" datetime,
		"deleted_at" datetime,
		"by" varchar(255),
		"msg" varchar(255),
		"at" datetime
	);
	`
	r.db.MustExec(stmt)
}

func (r *ReminderStore) Set(reminder Reminder) error {
	stmt := `INSERT INTO reminders("by", "msg", "at", "created_at", "updated_at") VALUES(?,?,?,?,?);`
	_, err := r.db.Exec(stmt, reminder.By, reminder.Msg, reminder.At, time.Now(), time.Now())
	return err
}

func (r *ReminderStore) Reminders() []Reminder {
	reminders := []Reminder{}
	r.db.Select(&reminders, `SELECT * FROM reminders WHERE "deleted_at" IS NULL ORDER BY "at" ASC;`)
	return reminders
}

func (r *ReminderStore) Delete(id uint) {
	r.db.Exec(`UPDATE reminders SET "deleted_at" = ? WHERE id = ?;`, time.Now(), id)
}
