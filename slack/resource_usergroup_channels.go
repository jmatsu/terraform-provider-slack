package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

func resourceSlackUserGroupChannels() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSlackUserGroupChannelsRead,
		CreateContext: resourceSlackUserGroupChannelsCreate,
		UpdateContext: resourceSlackUserGroupChannelsUpdate,
		DeleteContext: resourceSlackUserGroupChannelsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set("usergroup_id", d.Id())
				return schema.ImportStatePassthroughContext(ctx, d, m)
			},
		},

		Schema: map[string]*schema.Schema{
			"usergroup_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"channels": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func configureSlackUserGroupChannels(ctx context.Context, logger *Logger, d *schema.ResourceData, userGroup slack.UserGroup) {
	d.SetId(userGroup.ID)
	_ = d.Set("channels", append(userGroup.Prefs.Channels, userGroup.Prefs.Groups...))

	logger.debug(ctx, "Configured usergroups' default channels")
}

func resourceSlackUserGroupChannelsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	usergroupId := d.Get("usergroup_id").(string)

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": usergroupId,
	})

	logger.trace(ctx, "Start creating the default channels")

	iChannels := d.Get("channels").(*schema.Set).List()
	channelsIds := make([]string, len(iChannels))
	for i, v := range iChannels {
		channelsIds[i] = v.(string)
	}

	params := &slack.UserGroup{
		ID: usergroupId,
		Prefs: slack.UserGroupPrefs{
			Channels: channelsIds,
		},
	}

	userGroup, err := client.UpdateUserGroupContext(ctx, *params)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't add the default channels to the slack usergroup (%s)", usergroupId),
				Detail:   err.Error(),
			},
		}
	}

	configureSlackUserGroupChannels(ctx, logger, d, userGroup)

	return nil
}

func resourceSlackUserGroupChannelsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	currentId := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": currentId,
	})

	usergroupId := d.Get("usergroup_id").(string)

	logger.trace(ctx, "Start reading default channels of the usergroup")

	if usergroupId != d.Id() {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("it's not allowed to change usergroup id (from %s to %s)", currentId, usergroupId),
				Detail:   "Please move the state or create another resource instead",
			},
		}
	}

	// Use a cache for usergroups api call because the limitation is strict
	var userGroups *[]slack.UserGroup

	if !restoreJsonCache(userGroupListCacheFileName, &userGroups) {
		tempUserGroups, err := client.GetUserGroupsContext(ctx, func(params *slack.GetUserGroupsParams) {
			params.IncludeUsers = false
			params.IncludeCount = false
			params.IncludeDisabled = true
		})

		if err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Slack provider couldn't read the default channels of the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
					Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.list"),
				},
			}
		} else {
			logger.trace(ctx, "Got a response from Slack API")
		}

		userGroups = &tempUserGroups

		saveCacheAsJson(userGroupListCacheFileName, &userGroups)
	}

	if userGroups == nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Serious error happened while reading default channles of the slack usergroup (%s)", usergroupId),
				Detail:   "Internal provider error. Please open an issue at https://github.com/jmatsu/terraform-provider-slack",
			},
		}
	}

	for _, userGroup := range *userGroups {
		if userGroup.ID == usergroupId {
			configureSlackUserGroupChannels(ctx, logger, d, userGroup)
			return nil
		}
	}

	return nil
}

func resourceSlackUserGroupChannelsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	currentId := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": currentId,
	})

	logger.trace(ctx, "Start updating default channels of the usergroup")

	usergroupId := d.Get("usergroup_id").(string)

	if usergroupId != d.Id() {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("it's not allowed to change usergroup id (from %s to %s)", currentId, usergroupId),
				Detail:   "Please move the state or create another resource instead",
			},
		}
	}

	iChannels := d.Get("channels").(*schema.Set).List()
	channelsIds := make([]string, len(iChannels))
	for i, v := range iChannels {
		channelsIds[i] = v.(string)
	}

	params := &slack.UserGroup{
		ID: usergroupId,
		Prefs: slack.UserGroupPrefs{
			Channels: channelsIds,
		},
	}

	userGroup, err := client.UpdateUserGroupContext(ctx, *params)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't update the default channels of the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.update"),
			},
		}
	}

	configureSlackUserGroupChannels(ctx, logger, d, userGroup)

	return nil
}

func resourceSlackUserGroupChannelsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	currentId := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": currentId,
	})

	logger.trace(ctx, "Start destroying default channels of the usergroup")

	usergroupId := d.Get("usergroup_id").(string)

	if usergroupId != currentId {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("it's not allowed to change usergroup id (from %s to %s)", currentId, usergroupId),
				Detail:   "Please move the state or create another resource instead",
			},
		}
	}

	params := &slack.UserGroup{
		ID: usergroupId,
		Prefs: slack.UserGroupPrefs{
			Channels: []string{},
		},
	}

	// 0 default channels are allowed by spec
	if _, err := client.UpdateUserGroupContext(ctx, *params); err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't remove all default channels from the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.update"),
			},
		}
	}

	d.SetId("")

	logger.debug(ctx, "Cleared the resource id of this usergourps' default channels resource so it's going to be removed from the state")

	return nil
}
