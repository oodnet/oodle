package store

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Letter struct {
	ID        uint       `db:"id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
	From      string     `db:"from"`
	To        string     `db:"to"`
	Body      string     `db:"body"`
	When      time.Time  `db:"when"`
}

type TellStore struct {
	db *sqlx.DB
}

func NewTellStore(db *sqlx.DB) *TellStore {
	t := &TellStore{
		db: db,
	}
	t.Migrate()
	return t
}

func (t *TellStore) Migrate() {
	stmt := `
	CREATE TABLE IF NOT EXISTS "letters" (
		"id" integer primary key autoincrement,
		"created_at" datetime,
		"updated_at" datetime,
		"deleted_at" datetime,
		"from" varchar(255),
		"to" varchar(255),
		"body" varchar(255),
		"when" datetime
	);
	`
	t.db.MustExec(stmt)
}

func (t *TellStore) Send(l Letter) error {
	stmt := `INSERT INTO letters("from", "to", "body", "when", "created_at", "updated_at") VALUES(?,?,?,?,?,?);`
	_, err := t.db.Exec(stmt, l.From, l.To, l.Body, l.When, time.Now(), time.Now())
	return err
}

func (t *TellStore) GetLetters(to string) []*Letter {
	letters := []*Letter{}
	fmt.Println(t.db.Select(&letters, `SELECT * FROM letters WHERE "to" = ? AND "deleted_at" IS NULL;`, to))
	return letters
}

func (t *TellStore) Delete(letters []*Letter) {
	tx := t.db.MustBegin()
	for _, letter := range letters {
		tx.Exec(`UPDATE letters SET "deleted_at" = ? WHERE id = ?;`, time.Now(), letter.ID)
	}
	tx.Commit()
}

func (t *TellStore) HasMail(to string) bool {
	exists := false
	stmt := `SELECT EXISTS(SELECT * FROM letters WHERE "to" = ? AND "deleted_at" IS NULL);`
	t.db.Get(&exists, stmt, to)
	return exists
}
