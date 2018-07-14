package reputation

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// RepStore stores user reputations
type RepStore struct {
	db *sqlx.DB
}

func NewRepStore(db *sql.DB) *RepStore {
	rs := &RepStore{
		db: sqlx.NewDb(db, "sqlite3"),
	}
	rs.Migrate()
	return rs
}

// Migrate initializes the schema
func (r *RepStore) Migrate() {
	r.db.MustExec(`
		CREATE TABLE IF NOT EXISTS "user_reputations" (
			"id" integer primary key autoincrement,
			"giver" varchar(255) NOT NULL,
			"reciver" varchar(255) NOT NULL,
			"point" integer DEFAULT 0,
			"timestamp" datetime NOT NULL
		);`)
}

// LastGiven returns when/which point they last gave
func (r *RepStore) LastGiven(user string) (int, time.Time) {
	var point int
	var lastGiven time.Time
	r.db.QueryRow(`SELECT point, timestamp FROM 'user_reputations' WHERE giver = ? ORDER BY timestamp DESC LIMIT 1;`, user).
		Scan(&point, &lastGiven)
	return point, lastGiven
}

// Points returns total number of points for a given user
func (r *RepStore) Points(user string) int {
	points := 0
	r.db.QueryRow(`SELECT SUM(point) FROM 'user_reputations' WHERE reciver = ?;`, user).
		Scan(&points)
	return points
}

// Give stores a rep event into db and returns final points of the reciver
func (r *RepStore) Give(giver, reciver string, point int) (int, error) {
	tx := r.db.MustBegin()
	reciverPoints := 0
	tx.Exec(`
		INSERT INTO user_reputations(giver, reciver, point, 'timestamp') VALUES(?,?,?,?);`,
		giver, reciver, point, time.Now())
	tx.QueryRow(`SELECT SUM(point) FROM 'user_reputations' WHERE reciver = ?;`, reciver).
		Scan(&reciverPoints)
	return reciverPoints, tx.Commit()
}
