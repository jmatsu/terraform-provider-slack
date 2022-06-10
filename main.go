package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/remerge/terraform-provider-slack/slack"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: slack.Provider})
}
