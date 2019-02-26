package irc

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/cenkalti/backoff"
	"github.com/godwhoa/oodle/oodle"
	"github.com/lrstanley/girc"
	"github.com/sirupsen/logrus"
)

// Custom Dialer which prioritizes ipv4
type dialer struct{}

func (d *dialer) Dial(network, address string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp4", address)
	if err == nil {
		return
	}
	conn, err = net.Dial("tcp6", address)
	if err == nil {
		return
	}
	conn, err = net.Dial(network, address)
	return
}

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
	whoisreq  chan girc.Event
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
		cfg:      config,
		nick:     config.Nick,
		channel:  config.Channel,
		log:      log,
		whoisreq: make(chan girc.Event),
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
	client.Handlers.Add(girc.ALL_EVENTS, func(c *girc.Client, e girc.Event) {
		if e.Command == girc.PRIVMSG {
			fmt.Printf("PRIVMSG %s: %s\n", e.Source.Name, e.Trailing)
		}
		if e.Command == girc.NOTICE {
			fmt.Printf("NOTICE %s: %s\n", e.Source.Name, e.Trailing)
		}
		// Print error responses
		if len(e.Command) == 3 && strings.HasPrefix(e.Command, "4") {
			fmt.Printf("ERR_RPL: %s %s\n", e.Source.String(), e.Trailing)
		}
	})
	irc.client = client

	err := client.DialerConnect(&dialer{})
	if _, ok := err.(*girc.ErrInvalidConfig); ok || !irc.cfg.Retry {
		return err
	}

	return backoff.RetryNotify(
		func() error {
			return client.DialerConnect(&dialer{})
		},
		backoff.NewExponentialBackOff(),
		func(err error, dur time.Duration) {
			irc.log.Warnf("Connection failed with err: %s", err)
			irc.log.Warnf("Retrying in %s", dur)
		},
	)
}

// try to reclaim nick every 1 min until it is reclaimed
func (irc *IRCClient) reclaimNick() {
	for {
		if irc.cfg.Nick == irc.client.GetNick() {
			irc.log.Debugf("Nick reclaimed! Stoping loop")
			break
		}

		irc.client.Cmd.Whois(irc.cfg.Nick)
		if e := <-irc.whoisreq; e.Command == girc.ERR_NOSUCHNICK {
			irc.client.Cmd.Nick(irc.cfg.Nick)
			irc.log.Debugf("Sent /nick %s", irc.cfg.Nick)
		} else {
			irc.log.Debug("User with desired nick still connected.")
		}

		// Lets not get banned.
		time.Sleep(30 * time.Second)
	}
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

	if irc.cfg.Nick != irc.client.GetNick() {
		go irc.reclaimNick()
	}
	c.Cmd.Join(irc.cfg.Channel)
}

// Simplifies and sends the events
func (irc *IRCClient) onAll(c *girc.Client, e girc.Event) {
	if e.IsFromUser() || e.Source == nil || e.Source.Name == irc.cfg.Nick {
		return
	}
	nick, msg := e.Source.Name, e.Trailing
	switch e.Command {
	case girc.RPL_WHOISUSER, girc.ERR_NOSUCHNICK:
		irc.whoisreq <- e
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

// InChannel checks if a user is in the channel
func (irc *IRCClient) InChannel(nick string) bool {
	user := irc.client.LookupUser(nick)
	return user != nil && user.InChannel(irc.cfg.Channel)
}

// IsAdmin checks if a user is in the channel
func (irc *IRCClient) IsAdmin(nick string) bool {
	user := irc.client.LookupUser(nick)
	if user == nil || !user.InChannel(irc.cfg.Channel) {
		return false
	}
	perms, ok := user.Perms.Lookup(irc.cfg.Channel)
	return ok && perms.IsAdmin()
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

// Send sends an msg to the configured channel
func (irc *IRCClient) Send(message string) {
	irc.client.Cmd.Message(irc.cfg.Channel, message)
}

// SendTo sends to a specific user
func (irc *IRCClient) SendTo(user, message string) {
	irc.client.Cmd.Message(user, message)
}

// /msg hello world -> "/msg", ["##oodle-test", "hello", "world"]
func parseCmd(input string) (string, []string) {
	if !strings.HasPrefix(input, "/") {
		return "", []string{}
	}
	args := strings.Split(input, " ")
	if len(args) < 2 {
		return "", []string{}
	}
	return args[0], args[1:]
}

// MiniClient starts a mini clients which provides basic irc commands
func (irc *IRCClient) MiniClient() {

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		cmd, args := parseCmd(input)
		switch cmd {
		case "/msg":
			irc.SendTo(args[0], strings.Join(args[1:], " "))
		case "/me":
			irc.client.Cmd.Action(irc.cfg.Channel, strings.Join(args, " "))
		case "/nick":
			irc.client.Cmd.Nick(args[0])
		}

	}

	if err := scanner.Err(); err != nil {
		irc.log.Error(err)
	}
}
