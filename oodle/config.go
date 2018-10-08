package oodle

import "github.com/spf13/viper"

func SetDefaults() {
	viper.SetDefault("nick", "oodle-dev")
	viper.SetDefault("name", "oodle-dev")
	viper.SetDefault("server", "chat.freenode.net")
	viper.SetDefault("server", "chat.freenode.net")
	viper.SetDefault("port", 6667)
	viper.SetDefault("channel", "##oodle-test")
	viper.SetDefault("retry", true)
	viper.SetDefault("dbpath", "store.sqlite")
	viper.SetDefault("webhook_addr", "")
}
