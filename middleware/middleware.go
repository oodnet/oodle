package middleware

import (
	"time"

	"github.com/godwhoa/oodle/oodle"
)

type Middleware func(oodle.Command) oodle.Command

// Chain lets you chain multiple middleware
func Chain(cmd oodle.Command, middlewares ...Middleware) oodle.Command {
	if len(middlewares) == 0 {
		return cmd
	}

	// Wrap the first middleware with the command
	c := middlewares[len(middlewares)-1](cmd)
	// Wrap that with the rest of the middleware chain
	for i := len(middlewares) - 2; i >= 0; i-- {
		c = middlewares[i](c)
	}

	return c
}

func MinArg(l int) Middleware {
	fn := func(cmd oodle.Command) oodle.Command {
		next := cmd.Fn
		cmd.Fn = func(nick string, args []string) (reply string, err error) {
			if len(args) < l {
				return "", oodle.ErrUsage
			}
			return next(nick, args)
		}
		return cmd
	}
	return Middleware(fn)
}

func AdminOnly(check oodle.Checker) Middleware {
	fn := func(cmd oodle.Command) oodle.Command {
		next := cmd.Fn
		cmd.Fn = func(nick string, args []string) (reply string, err error) {
			if !check.IsAdmin(nick) {
				return "This command can only be executed by admins.", nil
			}
			return next(nick, args)
		}
	}
	return Middleware(fn)
}

func RegisteredOnly(check oodle.Checker) Middleware {
	fn := func(cmd oodle.Command) oodle.Command {
		next := cmd.Fn
		cmd.Fn = func(nick string, args []string) (reply string, err error) {
			if !check.IsRegistered(nick) {
				return "This command can only be executed by registered users.", nil
			}
			return next(nick, args)
		}
		return cmd
	}
	return Middleware(fn)
}

func RateLimit(limit int, duration time.Duration) Middleware {
	executions := 0
	last := time.Now()
	fn := func(cmd oodle.Command) oodle.Command {
		next := cmd.Fn
		cmd.Fn = func(nick string, args []string) (reply string, err error) {
			if executions > limit {
				return "Rate limited.", nil
			}
			if time.Since(last) >= duration {
				executions = 0
			}
			executions++
			last = time.Now()
			return next(nick, args)
		}
		return cmd
	}
	return Middleware(fn)
}
