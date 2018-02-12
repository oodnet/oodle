package oodle

type Config struct {
	Nick        string `toml:"nick"`
	Name        string `toml:"name"`
	User        string `toml:"user"`
	Server      string `toml:"server"`
	Port        int    `toml:"port"`
	Channel     string `toml:"channel"`
	Retry       bool   `toml:"retry"`
	SASLUser    string `toml:"sasl_user"`
	SASLPass    string `toml:"sasl_pass"`
	DBPath      string `toml:"dbpath"`
	WebHookAddr string `toml:"webhook_addr"`
	Secret      string `toml:"secret"`
}
