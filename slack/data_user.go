package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

const (
	userListCacheFileName = "users.json"
	userQueryTypeID       = "id"
	userQueryTypeName     = "name"
	userQueryTypeEmail    = "email"
)

func dataSourceSlackUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSlackUserRead,

		Schema: map[string]*schema.Schema{
			"query_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateEnums([]string{userQueryTypeID, userQueryTypeName, userQueryTypeEmail}),
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

func dataSourceSlackUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	queryType := d.Get("query_type").(string)
	queryValue := d.Get("query_value").(string)

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"data":        "slack_user",
		"query_type":  queryType,
		"query_value": queryValue,
	})

	configureUserFunc := func(d *schema.ResourceData, user slack.User) {
		d.SetId(user.ID)
		_ = d.Set("name", user.Name)
		_ = d.Set("real_name", user.RealName)
		_ = d.Set("is_admin", user.IsAdmin)
		_ = d.Set("is_owner", user.IsOwner)
		_ = d.Set("is_bot", user.IsBot)
		_ = d.Set("has_2fa", user.Has2FA)

		logger.debug(ctx, "User #%s (isBot = %t)", d.Get("name").(string), d.Get("is_bot").(bool))
	}

	if queryType == userQueryTypeID {
		logger.trace(ctx, "Start reading the slack user by user_id")

		// https://api.slack.com/docs/rate-limits#tier_t4
		user, err := client.GetUserInfoContext(ctx, queryValue)

		if err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Slack provider couldn't find a slack user (%s) due to *%s*", queryValue, err.Error()),
					Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/users.info"),
				},
			}
		} else {
			logger.trace(ctx, "Got a response from Slack api")
		}

		configureUserFunc(d, *user)
		return nil
	}

	if queryType == userQueryTypeEmail {
		logger.trace(ctx, "Start reading the slack user by email")

		// https://api.slack.com/methods/users.lookupByEmail
		// https://api.slack.com/docs/rate-limits#tier_t3
		user, err := client.GetUserByEmailContext(ctx, queryValue)

		if err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Slack provider couldn't find a slack user (%s) due to *%s*", queryValue, err.Error()),
					Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/users.lookupByEmail"),
				},
			}
		} else {
			logger.trace(ctx, "Got a response from Slack api")
		}

		configureUserFunc(d, *user)
		return nil
	}

	if queryType == userQueryTypeName {
		logger.trace(ctx, "Start reading the slack user by user_name")

		// Use a cache for users api call because the limitation is stricter than user.info
		var users *[]slack.User

		if !restoreJsonCache(userListCacheFileName, &users) {
			tempUsers, err := client.GetUsersContext(ctx)

			if err != nil {
				return diag.Diagnostics{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("Slack provider couldn't find a slack user (%s) due to *%s*", queryValue, err.Error()),
						Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/users.list"),
					},
				}
			} else {
				logger.trace(ctx, "Got a response from Slack api")
			}

			users = &tempUsers

			saveCacheAsJson(userListCacheFileName, &users)
		}

		if users == nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Serious error happened while finding a slack user (%s)", queryValue),
					Detail:   "Please use another query_type or open an issue at https://github.com/jmatsu/terraform-provider-slack",
				},
			}
		}

		for _, user := range *users {
			if dataSourceSlackUserMatch(&user, queryType, queryValue) {
				logger.debug(ctx, "Found a user")

				configureUserFunc(d, user)
				return nil
			}
		}

		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't find a slack user (%s)", queryValue),
				Detail:   fmt.Sprintf("In general, slack username is not reliable and non-determistic. It's better to use %s or %s to look up a user instead and so we've deprecated this query type actually", userQueryTypeEmail, userQueryTypeID),
			},
		}
	}

	logger.error(ctx, "this flow is an unexpected")

	return diag.Diagnostics{
		{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s in query_type is not acceptable", queryType),
			Detail:   fmt.Sprintf("Either one of %s, %s and %s is allowed", userQueryTypeID, userQueryTypeEmail, userQueryTypeID),
		},
	}
}

func dataSourceSlackUserMatch(user *slack.User, queryType string, queryValue string) bool {
	switch queryType {
	case userQueryTypeName:
		return user.Name == queryValue || user.RealName == queryValue || user.Profile.DisplayName == queryValue
	case userQueryTypeEmail:
		return user.Profile.Email == queryValue
	case userQueryTypeID:
		return user.ID == queryValue
	}
	return false
}
