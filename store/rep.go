package store

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type UserRep struct {
	ID     uint      `db:"id"`
	User   string    `db:"user"`
	Points int       `db:"points"`
	Next   time.Time `db:"next"`
}

func NewRepStore(db *sqlx.DB, cooldowns map[int]time.Duration) *RepStore {
	store := &RepStore{db: db, cooldowns: cooldowns}
	store.Migrate()
	return store
}

type RepStore struct {
	db        *sqlx.DB
	cooldowns map[int]time.Duration
}

func (r *RepStore) Migrate() {
	stmt := `
	CREATE TABLE IF NOT EXISTS "user_reputations" (
		"id" integer primary key autoincrement,
		"user" varchar(255) NOT NULL UNIQUE,
		"points" integer DEFAULT 0,
		"next" datetime NOT NULL
	)
	`
	r.db.MustExec(stmt)
}

// Initializes a user if they don't exist
func initUser(tx *sqlx.Tx, user string) {
	tx.Exec(`INSERT OR IGNORE INTO user_reputations(user, next) VALUES(?,?);`, user, time.Now().Add(-1*time.Second))
}

func (r *RepStore) Inc(giver, reciver string, point int) error {
	cooldown := r.cooldowns[point]
	next := time.Now().Add(cooldown)
	tx := r.db.MustBegin()
	// initialize giver/reciver if they don't exist
	initUser(tx, giver)
	initUser(tx, reciver)
	// increment rep for reciver
	if _, err := tx.Exec(`UPDATE user_reputations SET points = points + ? WHERE user = ?;`, point, reciver); err != nil {
		return err
	}
	// update next for giver
	if _, err := tx.Exec(`UPDATE user_reputations SET next = ? WHERE user = ?;`, next, giver); err != nil {
		return err
	}
	return tx.Commit()
}

// GetUserRep gets a UserRep by nick
func (r *RepStore) GetUserRep(user string) (*UserRep, error) {
	userRep := &UserRep{}
	tx := r.db.MustBegin()
	initUser(tx, user)
	tx.Get(userRep, `SELECT * FROM user_reputations WHERE user = ?;`, user)
	return userRep, tx.Commit()
}
