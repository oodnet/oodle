package irc

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/cenkalti/backoff"
	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server   string
	Port     int
	Channel  string
	Nick     string
	Name     string
	User     string
	SASLUser string
	SASLPass string
	Retry    bool
}

// IRCClient is a thin wrapper around girc client
type IRCClient struct {
	nick      string
	channel   string
	cfg       *Config
	callbacks []func(event interface{})
	client    *girc.Client
	log       *logrus.Logger
}

func NewIRCClient(log *logrus.Logger) *IRCClient {

	config := &Config{
		Server:   viper.GetString("server"),
		Port:     viper.GetInt("port"),
		Channel:  viper.GetString("channel"),
		Nick:     viper.GetString("nick"),
		Name:     viper.GetString("name"),
		User:     viper.GetString("user"),
		SASLUser: viper.GetString("sasl_user"),
		SASLPass: viper.GetString("sasl_pass"),
		Retry:    viper.GetBool("retry"),
	}

	return &IRCClient{
		cfg:     config,
		nick:    config.Nick,
		channel: config.Channel,
		log:     log,
	}
}

func (irc *IRCClient) Connect() error {
	gircConf := girc.Config{
		Server:      irc.cfg.Server,
		Port:        irc.cfg.Port,
		Nick:        irc.cfg.Nick,
		User:        irc.cfg.User,
		Name:        irc.cfg.Name,
		RecoverFunc: func(_ *girc.Client, e *girc.HandlerError) { irc.log.Errorf("%+v", e) },
	}
	if irc.cfg.SASLUser != "" && irc.cfg.SASLPass != "" {
		gircConf.SASL = &girc.SASLPlain{User: irc.cfg.SASLUser, Pass: irc.cfg.SASLPass}
	}

	client := girc.New(gircConf)
	client.Handlers.Add(girc.ALL_EVENTS, irc.onAll)
	client.Handlers.Add(girc.CONNECTED, irc.onConnect)
	irc.client = client

	err := client.Connect()
	if _, ok := err.(*girc.ErrInvalidConfig); ok || !irc.cfg.Retry {
		return err
	}
	return backoff.RetryNotify(client.Connect, backoff.NewExponentialBackOff(), func(err error, dur time.Duration) {
		irc.log.Warnf("Connection failed with err: %s", err)
		irc.log.Warnf("Retrying in %s", dur)
	})
}

// Close closes the connection
func (irc *IRCClient) Close() {
	irc.client.Close()
}

func (irc *IRCClient) onConnect(c *girc.Client, e girc.Event) {
	irc.log.WithFields(logrus.Fields{
		"server":  irc.cfg.Server,
		"port":    irc.cfg.Port,
		"channel": irc.cfg.Channel,
		"nick":    c.GetNick(),
	}).Info("Connected!")
	c.Cmd.Join(irc.cfg.Channel)
}

// Simplifies and sends the events
func (irc *IRCClient) onAll(c *girc.Client, e girc.Event) {
	if e.IsFromUser() || e.Source == nil || e.Source.Name == irc.cfg.Nick {
		return
	}
	nick, msg := e.Source.Name, e.Trailing
	switch e.Command {
	case girc.JOIN:
		irc.sendEvent(oodle.Join{Nick: nick})
	case girc.PART:
		irc.sendEvent(oodle.Leave{Nick: nick})
	case girc.PRIVMSG:
		irc.sendEvent(oodle.Message{Nick: nick, Msg: msg})
	}
}

// Sends an event to all callbacks
func (irc *IRCClient) sendEvent(event interface{}) {
	for _, callback := range irc.callbacks {
		callback(event)
	}
}

// OnEvent registers a callback
func (irc *IRCClient) OnEvent(callback func(event interface{})) {
	irc.callbacks = append(irc.callbacks, callback)
}

// Send sends an msg to the configured channel
func (irc *IRCClient) Send(message string) {
	irc.client.Cmd.Message(irc.cfg.Channel, message)
}

// InChannel checks if a user is in the channel
func (irc *IRCClient) InChannel(nick string) bool {
	user := irc.client.LookupUser(nick)
	return user != nil && user.InChannel(irc.cfg.Channel)
}

// IsRegistered checks if a user is registered
func (irc *IRCClient) IsRegistered(nick string) bool {
	user := irc.client.LookupUser(nick)
	return user != nil && user.Extras.Account != ""
}

// Sendf works like printf but for irc msgs.
func (irc *IRCClient) Sendf(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	for _, msg := range strings.Split(message, "\n") {
		irc.client.Cmd.Message(irc.cfg.Channel, msg)
	}
}
