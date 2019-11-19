package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nlopes/slack"
	"log"
)

const userListCacheFileName = "users.json"

func dataSourceSlackUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSlackUserRead,

		Schema: map[string]*schema.Schema{
			"query_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateEnums([]string{"id", "name"}),
			},
			"query_value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"real_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_admin": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_owner": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_bot": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"has_2fa": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceSlackUserRead(d *schema.ResourceData, meta interface{}) error {
	queryType := d.Get("query_type").(string)
	queryValue := d.Get("query_value").(string)

	configureUserFunc := func(d *schema.ResourceData, user slack.User) {
		d.SetId(user.ID)
		_ = d.Set("name", user.Name)
		_ = d.Set("real_name", user.RealName)
		_ = d.Set("is_admin", user.IsAdmin)
		_ = d.Set("is_owner", user.IsOwner)
		_ = d.Set("is_bot", user.IsBot)
		_ = d.Set("has_2fa", user.Has2FA)
	}

	log.Printf("[INFO] Refreshing Slack User: %s", queryValue)

	client := meta.(*Team).client
	ctx := context.WithValue(context.Background(), ctxId, queryValue)

	if queryType == "id" {
		// https://api.slack.com/docs/rate-limits#tier_t4
		user, err := client.GetUserInfoContext(ctx, queryValue)

		if err != nil {
			return err
		}

		configureUserFunc(d, *user)
		return nil
	}

	// Use a cache for users api call because the limitation is more strict than user.info
	var users *[]slack.User

	if !restoreJsonCache(userListCacheFileName, &users) {
		tempUsers, err := client.GetUsersContext(ctx)

		if err != nil {
			return err
		}

		users = &tempUsers

		saveCacheAsJson(userListCacheFileName, &users)
	}

	if users == nil {
		panic(fmt.Errorf("a serious error happened. please create an issue to https://github.com/jmatsu/terraform-provider-slack"))
	}

	for _, user := range *users {
		if user.Name == queryValue || user.RealName == queryValue || user.Profile.DisplayName == queryValue {
			configureUserFunc(d, user)
			return nil
		}
	}

	return fmt.Errorf("a slack user (%s) is not found", queryValue)
}
