package slack

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	var p *schema.Provider
	p = &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SLACK_TOKEN", nil),
				Description: descriptions["token"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"slack_user":      dataSourceSlackUser(),
			"slack_usergroup": dataSourceUserGroup(),
			"slack_channel":   dataSourceChannel(),
			"slack_group":     dataSourceGroup(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"slack_usergroup":         resourceSlackUserGroup(),
			"slack_usergroup_members": resourceSlackUserGroupMembers(),
			"slack_channel":           resourceSlackChannel(),
			"slack_group":             resourceSlackGroup(),
		},
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token": "The OAuth token used to connect to Slack.",
	}
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		config := Config{
			Token: d.Get("token").(string),
		}

		meta, err := config.Client()
		if err != nil {
			return nil, err
		}

		meta.(*Team).StopContext = p.StopContext()

		return meta, nil
	}
}
