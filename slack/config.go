package slack

import (
	"context"
	"errors"

	"github.com/slack-go/slack"
)

const (
	ctxId = 1
)

type Config struct {
	Token string
}

type Team struct {
	client      *slack.Client
	auth        *slack.AuthTestResponse
	StopContext context.Context
}

func (c *Config) Client() (interface{}, error) {
	client := slack.New(c.Token)
	auth, err := client.AuthTest()
	if err != nil {
		return nil, errors.New("Could not authorize with given token")
	}

	return &Team{
		client: client,
		auth:   auth,
	}, nil
}
