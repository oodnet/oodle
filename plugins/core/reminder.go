package core

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/godwhoa/oodle/oodle"
	"github.com/jmoiron/sqlx"
)

/*
> Persist reminders to database
> A loop queries the database for reminders which need to be sent out every 1sec
> If user is in channel then send it out
> If not add it to .tell's store
*/

var durationRegex = regexp.MustCompile(`([0-9]+)|month|day|hour|min|sec|mon|d|h|m|s`)

func inOrder(units []time.Duration) bool {
	if len(units) == 0 || len(units) == 1 {
		return true
	}
	for i := 1; i < len(units); i++ {
		if units[i-1] < units[i] {
			return false
		}
	}
	return true
}

func dup(units []time.Duration) bool {
	if len(units) == 0 || len(units) == 1 {
		return false
	}
	set := make(map[time.Duration]struct{})
	for _, unit := range units {
		if _, ok := set[unit]; ok {
			return true
		}
		set[unit] = struct{}{}
	}
	return false
}

// converts 1d1hour5min to a time.Duration obj
func parseDuration(format string) (time.Duration, error) {
	multiplier := []int{}
	units := []time.Duration{}
	fmt.Println(durationRegex.MatchString(format))
	matches := durationRegex.FindAllString(format, -1)
	for _, m := range matches {
		i, err := strconv.Atoi(m)
		if err == nil {
			multiplier = append(multiplier, i)
		}
		switch m {
		case "month", "mon":
			units = append(units, time.Hour*24*30)
		case "day", "d":
			units = append(units, time.Hour*24)
		case "hour", "h":
			units = append(units, time.Hour*1)
		case "min", "m":
			units = append(units, time.Minute*1)
		case "sec", "s":
			units = append(units, time.Second*1)
		}
	}
	if len(multiplier) != len(units) {
		return 0 * time.Second, fmt.Errorf("Invalid duration format")
	}
	if !inOrder(units) {
		return 0 * time.Second, fmt.Errorf("Units not in proper order")
	}
	if dup(units) {
		return 0 * time.Second, fmt.Errorf("Repeating units in format")
	}
	duration := 0 * time.Second
	for i := 0; i < len(units); i++ {
		duration += time.Duration(multiplier[i]) * units[i]
	}
	return duration, nil
}

type RemindIn struct {
	checker oodle.Checker
	sender  oodle.Sender
	store   *ReminderStore
	mailbox *MailBox
}

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
	if len(args) < 2 {
		return "", oodle.ErrUsage
	}
	duration, err := parseDuration(args[0])
	if err != nil {
		return err.Error(), nil
	}
	msg := strings.Join(args[1:], " ")
	r.store.Set(Reminder{
		By:  nick,
		Msg: msg,
		At:  time.Now().Add(duration),
	})
	return "Reminder set!", nil
}

func (r *RemindIn) sendout() {
	reminders := r.store.Reminders()
	for _, reminder := range reminders {
		fmt.Println(reminder.At.Sub(time.Now()))
		if time.Now().After(reminder.At) {
			r.send(reminder)
			r.store.Delete(reminder.ID)
		}
	}
}

// Watch watches the store and sends out reminder that need to be out
func (r *RemindIn) Watch() {
	for {
		r.sendout()
		time.Sleep(1 * time.Second)
	}
}

func (r *RemindIn) Command() oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "remindin",
		Description: "Lets you set yourself a reminder",
		Usage:       ".reminder <duration> <msg>",
		Fn:          r.fn,
	}
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
	fmt.Println(r.db.Select(&reminders, `SELECT * FROM reminders WHERE "deleted_at" IS NULL ORDER BY "at" ASC;`))
	return reminders
}

func (r *ReminderStore) Delete(id uint) {
	r.db.Exec(`UPDATE reminders SET "deleted_at" = ? WHERE id = ?;`, time.Now(), id)
}
