package slack

import (
	"github.com/slack-go/slack"
)

const (
	ctxId = 1
)

type Config struct {
	Token string
}

type Team struct {
	client *slack.Client
}

func (c *Config) Client() (interface{}, error) {
	var team Team

	team.client = slack.New(c.Token)

	return &team, nil
}
