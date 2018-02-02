package oodle

type Config struct {
	Nick     string `toml:"nick"`
	Server   string `toml:"server"`
	Port     int    `toml:"port"`
	Channel  string `toml:"channel"`
	Retry    bool   `toml:"retry"`
	SASLUser string `toml:"sasl_user"`
	SASLPass string `toml:"sasl_pass"`
}
