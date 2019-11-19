package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nlopes/slack"
	"gopkg.in/djherbis/times.v1"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const userListCacheDir = "./.terraform/plugins/.cache/terraform-provider-slack"
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
	var users []slack.User
	var shouldApiCall = true
	var userListCacheFile string

	// if creating a directory fails, a cache system won't work but the processing should be executed
	_ = os.MkdirAll(userListCacheDir, 0755)
	userListCacheFile = strings.Join([]string{userListCacheDir, userListCacheFileName}, string(os.PathSeparator))

	// cache active duration is 1 min because api limitation is based on tier_t2
	if t, err := times.Stat(userListCacheFile); err == nil {
		if !time.Now().After(t.ModTime().Add(1 * time.Minute)) {
			if bytes, err := ioutil.ReadFile(userListCacheFile); err == nil {
				shouldApiCall = json.Unmarshal(bytes, &users) != nil
			}
		}
	}

	if shouldApiCall {
		var err error

		// https://api.slack.com/docs/rate-limits#tier_t2
		if users, err = client.GetUsersContext(ctx); err == nil {
			if cache, err := json.Marshal(users); err == nil {
				_ = ioutil.WriteFile(userListCacheFile, cache, 0644)
			}
		} else {
			return err
		}
	}

	for _, user := range users {
		if user.Name == queryValue || user.RealName == queryValue || user.Profile.DisplayName == queryValue {
			configureUserFunc(d, user)
			return nil
		}
	}

	return fmt.Errorf("a slack user (%s) is not found", queryValue)
}
