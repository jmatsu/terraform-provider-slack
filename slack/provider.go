package slack

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token": "The OAuth token used to connect to Slack.",
	}

	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string, commit string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"token": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("SLACK_TOKEN", nil),
					Description: descriptions["token"],
				},
			},

			DataSourcesMap: map[string]*schema.Resource{
				"slack_user":         dataSourceSlackUser(),
				"slack_usergroup":    dataSourceUserGroup(),
				"slack_conversation": dataSourceConversation(),
			},

			ResourcesMap: map[string]*schema.Resource{
				"slack_usergroup":          resourceSlackUserGroup(),
				"slack_usergroup_members":  resourceSlackUserGroupMembers(),
				"slack_conversation":       resourceSlackConversation(),
				"slack_usergroup_channels": resourceSlackUserGroupChannels(),
			},
		}

		p.ConfigureContextFunc = configureProvider(version, commit)

		return p
	}
}

func configureProvider(version string, commit string) schema.ConfigureContextFunc {
	return func(context context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		config := Config{
			Token: d.Get("token").(string),
		}

		meta, err := config.ProviderContext(version, commit)

		if err != nil {
			return nil, diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Cannot create a Slack client in set up",
					Detail:   err.Error(),
				},
			}
		}

		return meta, nil
	}
}
