package discord

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "strings"

	"github.com/jmoiron/sqlx"
	"github.com/lrstanley/girc"
	"github.com/oodnet/oodle/events"
	"github.com/oodnet/oodle/oodle"
	"github.com/spf13/viper"
)

type Message struct {
	Username  *string `json:"username,omitempty"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
	Content   *string `json:"content,omitempty"`
}

type Discord struct {
	checker  oodle.Checker
	sender   oodle.Sender
	store    *OptInStore
	queryNow chan bool
}

var OptInCache = []OptInUser{}

func Register(deps *oodle.Deps) error {
	bot, irc, db := deps.Bot, deps.IRC, deps.DB
	url := viper.GetString("discord_webhook")

	discord := &Discord{irc, irc, NewOptInStore(db), make(chan bool)}

	UpdateCache(discord.store)

	bot.RegisterCommands(OptIn(irc, discord))
	bot.RegisterCommands(OptOut(irc, discord))
	bot.RegisterTriggers(OnMessage(irc, irc, url, discord))

	return nil
}

func OnMessage(sender oodle.Sender, check oodle.Checker, url string, discord *Discord) oodle.Trigger {
	trigger := func(event girc.Event) {
		if !events.Is(event, events.MESSAGE) {
			return
		}
		nick, msg := events.Message(event)

		discord.store.Users()

		if !IsOptedIn(nick) {
			sender.Sendf("You haven't opted in!")
			return
		}

		RelayMessage(nick, msg, url)
	}

	return oodle.Trigger(trigger)
}

func OptIn(sender oodle.Sender, discord *Discord) oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "opt-in",
		Description: "Opt-in to the discord message bridge.",
		Usage:       "Simply type .opt-in to opt-in",
		Fn: func(nick string, arg []string) (reply string, err error) {
			_, new := discord.store.Set(OptInUser{Nick: nick})
			if new {
				return "You have opted in to have your messages relayed to Discord, type .opt-out to opt back out.", nil
			} else {
				return "You have already opted in to the discord bridge, type .opt-out to opt out.", nil
			}
		},
	}
}

func OptOut(sender oodle.Sender, discord *Discord) oodle.Command {
	return oodle.Command{
		Prefix:      ".",
		Name:        "opt-out",
		Description: "Opt-out of the discord message bridge.",
		Usage:       "Simply type .opt-out to opt-out",
		Fn: func(nick string, arg []string) (reply string, err error) {
			discord.store.Delete(OptInUser{Nick: nick})
			return "You have opted out of the Discord bridge, type .opt-in to opt back in.", nil
		},
	}
}

func SendWebhook(url string, message Message) error {
	payload := new(bytes.Buffer)

	err := json.NewEncoder(payload).Encode(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", payload)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		defer resp.Body.Close()

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf(string(responseBody))
	}

	return nil
}

func RelayMessage(nick string, content string, url string) {
	message := Message{
		Username: &nick,
		Content:  &content,
	}

	err := SendWebhook(url, message)
	if err != nil {
		log.Fatal(err)
	}
}

type OptInUser struct {
	ID   uint   `db:"id"`
	Nick string `db:"nick"`
}

type OptInStore struct {
	db *sqlx.DB
}

func NewOptInStore(db *sql.DB) *OptInStore {
	o := &OptInStore{
		db: sqlx.NewDb(db, "sqlite3"),
	}
	o.Migrate()
	return o
}

func (o *OptInStore) Migrate() {
	stmt := `
	CREATE TABLE IF NOT EXISTS "optin" (
		"id" integer primary key autoincrement,
		"nick" varchar(255)
	);
	`
	o.db.MustExec(stmt)
}

func (o *OptInStore) Set(user OptInUser) (error, bool) {
	users := []OptInUser{}
	o.db.Select(&users, `SELECT * FROM optin where nick = ?;`, user.Nick)

	if len(users) > 0 {
		log.Print("Skipping inserting users: ", user.Nick)
		return nil, false
	}

	log.Print("Inserting user: ", user.Nick)

	stmt := `INSERT INTO optin("nick") VALUES(?);`
	_, err := o.db.Exec(stmt, user.Nick)

	UpdateCache(o)

	return err, true
}

func (o *OptInStore) Users() []OptInUser {
	users := []OptInUser{}
	o.db.Select(&users, `SELECT * FROM optin;`)

	log.Print("Users in optin table:")
	for _, u := range users {
		log.Print(u.Nick)
	}

	UpdateCache(o)

	return users
}

func (o *OptInStore) Delete(user OptInUser) {
	log.Print("Deleting user from optin table: ", user.Nick)
	o.db.Exec(`DELETE FROM optin WHERE nick = ?;`, user.Nick)

	UpdateCache(o)
}

func UpdateCache(o *OptInStore) {
	OptInCache = nil
	o.db.Select(&OptInCache, `SELECT * FROM optin;`)
}

func IsOptedIn(nick string) bool {
	for _, u := range OptInCache {
		if u.Nick == nick {
			return true
		}
	}

	return false
}
