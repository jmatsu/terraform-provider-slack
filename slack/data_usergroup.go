package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nlopes/slack"
	"log"
)

func dataSourceUserGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSlackUserGroupRead,

		Schema: map[string]*schema.Schema{
			"usergroup_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"auto_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSlackUserGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	usergroupId := d.Get("usergroup_id").(string)
	ctx := context.WithValue(context.Background(), ctxId, usergroupId)

	log.Printf("[DEBUG] Reading usergroup: %s", usergroupId)
	groups, err := client.GetUserGroupsContext(ctx, func(params *slack.GetUserGroupsParams) {
		params.IncludeUsers = false
		params.IncludeCount = false
		params.IncludeDisabled = true
	})

	if err != nil {
		return err
	}

	for _, group := range groups {
		if group.ID == usergroupId {
			d.SetId(group.ID)
			_ = d.Set("handle", group.Handle)
			_ = d.Set("name", group.Name)
			_ = d.Set("description", group.Description)
			_ = d.Set("auto_type", group.AutoType)
			_ = d.Set("team_id", group.TeamID)
			return nil
		}
	}

	return fmt.Errorf("%s could not be found", usergroupId)
}
