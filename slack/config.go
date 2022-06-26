package slack

import (
	"github.com/slack-go/slack"
)

type Config struct {
	Token string
}

type Team struct {
	client *slack.Client
	logger *Logger
}

func (c *Config) ProviderContext(version string, commit string) (*Team, error) {
	var team Team

	team.client = slack.New(c.Token)
	team.logger = configureLogger(version, commit)

	return &team, nil
}
