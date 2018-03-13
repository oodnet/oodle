package oodle

import "time"

type Config struct {
	Nick           string            `toml:"nick"`
	Name           string            `toml:"name"`
	User           string            `toml:"user"`
	Server         string            `toml:"server"`
	Port           int               `toml:"port"`
	Channel        string            `toml:"channel"`
	Retry          bool              `toml:"retry"`
	SASLUser       string            `toml:"sasl_user"`
	SASLPass       string            `toml:"sasl_pass"`
	DBPath         string            `toml:"dbpath"`
	WebHookAddr    string            `toml:"webhook_addr"`
	Secret         string            `toml:"secret"`
	Commands       []string          `toml:"commands"`
	Points         []int             `toml:"points"`
	Cooldowns      []Duration        `toml:"cooldowns"`
	RegisteredOnly bool              `toml:"registered_only"`
	CustomCommands map[string]string `toml:"custom_commands"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
