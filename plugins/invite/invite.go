package invite

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/godwhoa/oodle/oodle"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Register(deps *oodle.Deps) error {
	sender, bot, log := deps.IRC, deps.Bot, deps.Logger
	schema := `
	CREATE TABLE IF NOT EXISTS tokens (
		id INT(6) UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		token VARCHAR(36) NOT NULL,
		nick VARCHAR(30) NOT NULL,
		created TIMESTAMP NOT NULL
		)
	`
	dsn := viper.GetString("token_mysql")
	if dsn == "" {
		return nil
	}
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	_, err = db.Exec(schema)
	bot.RegisterCommands(Invite(sender, log, db))
	return nil
}

func Invite(sender oodle.Sender, log *logrus.Logger, db *sqlx.DB) oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "invite",
		Description: "Generates an random invite code.",
		Usage:       ".invite",
		Fn: func(nick string, args []string) (string, error) {
			token, err := uuid.NewV4()
			if err != nil {
				return "Could not generate token", nil
			}
			_, err = db.Exec(`INSERT INTO tokens(token, nick, created) VALUES(?,?,?);`,
				token.String(), nick, time.Now().UTC())
			if err != nil {
				log.Debug(err)
				return "Database error.", nil
			}
			sender.SendTo(nick, token.String())
			return "Invite sent.", nil
		},
	}
}
