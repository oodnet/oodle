package bot

import (
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

type IRCClient struct {
	Server   string
	Port     int
	Channel  string
	Nick     string
	Name     string
	User     string
	SASLUser string
	SASLPass string
	Retry    bool

	callbacks []func(event interface{})
	client    *girc.Client
	log       *logrus.Logger
}

func NewIRCClient(log *logrus.Logger, conf *oodle.Config) *IRCClient {
	return &IRCClient{
		Server:   conf.Server,
		Port:     conf.Port,
		Channel:  conf.Channel,
		Nick:     conf.Nick,
		Name:     conf.Name,
		User:     conf.User,
		SASLUser: conf.SASLUser,
		SASLPass: conf.SASLPass,
		Retry:    conf.Retry,
		log:      log,
	}
}

func (irc *IRCClient) Connect() error {
	gircConf := girc.Config{
		Server:      irc.Server,
		Port:        irc.Port,
		Nick:        irc.Nick,
		User:        irc.User,
		Name:        irc.Name,
		RecoverFunc: func(_ *girc.Client, e *girc.HandlerError) { irc.log.Errorln(e.Error()) },
	}
	if irc.SASLUser != "" && irc.SASLPass != "" {
		gircConf.SASL = &girc.SASLPlain{User: irc.SASLUser, Pass: irc.SASLPass}
	}

	client := girc.New(gircConf)
	client.Handlers.Add(girc.ALL_EVENTS, irc.onAll)
	client.Handlers.Add(girc.CONNECTED, irc.onConnect)
	irc.client = client

	err := client.Connect()
	if _, ok := err.(*girc.ErrInvalidConfig); ok || !irc.Retry {
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
		"server":  irc.Server,
		"port":    irc.Port,
		"channel": irc.Channel,
		"nick":    c.GetNick(),
	}).Info("Connected!")
	c.Cmd.Join(irc.Channel)
}

// Simplifies and sends the events
func (irc *IRCClient) onAll(c *girc.Client, e girc.Event) {
	if e.Source == nil || e.Source.Name == irc.Nick {
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

// Sends and event to all callbacks
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
	for _, msg := range strings.Split(message, "\n") {
		irc.client.Cmd.Message(irc.Channel, msg)
	}
}
