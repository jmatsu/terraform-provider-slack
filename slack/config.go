package slack

import (
	"context"
	"github.com/nlopes/slack"
)

const (
	ctxId = 1
)

type Config struct {
	Token string
}

type Team struct {
	client      *slack.Client
	StopContext context.Context
}

func (c *Config) Client() (interface{}, error) {
	var team Team

	team.client = slack.New(c.Token)

	return &team, nil
}
