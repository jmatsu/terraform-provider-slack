package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/jmatsu/terraform-slack-provider/slack"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: slack.Provider})
}
