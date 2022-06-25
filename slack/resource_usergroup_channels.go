package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"log"
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

func configureSlackUserGroupChannels(d *schema.ResourceData, userGroup slack.UserGroup) {
	d.SetId(userGroup.ID)
	_ = d.Set("channels", append(userGroup.Prefs.Channels, userGroup.Prefs.Groups...))
}

func resourceSlackUserGroupChannelsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	usergroupId := d.Get("usergroup_id").(string)
	log.Printf("[DEBUG] Creating usergroup channels relation: %s", usergroupId)

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
				Summary:  fmt.Sprintf("provider cannot add the default channels to the slack usergroup (%s)", usergroupId),
				Detail:   err.Error(),
			},
		}
	}

	configureSlackUserGroupChannels(d, userGroup)

	return nil
}

func resourceSlackUserGroupChannelsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	currentId := d.Id()
	usergroupId := d.Get("usergroup_id").(string)
	log.Printf("[DEBUG] Reading usergroup channels relation: %s", usergroupId)

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
					Summary:  fmt.Sprintf("provider cannot read the default channels of the slack usergroup (%s)", usergroupId),
					Detail:   err.Error(),
				},
			}
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
			configureSlackUserGroupChannels(d, userGroup)
			return nil
		}
	}

	return nil
}

func resourceSlackUserGroupChannelsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	currentId := d.Id()
	usergroupId := d.Get("usergroup_id").(string)
	log.Printf("[DEBUG] Updating usergroup channels relation: %s", usergroupId)

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
				Summary:  fmt.Sprintf("provider cannot update the default channels of the slack usergroup (%s)", usergroupId),
				Detail:   err.Error(),
			},
		}
	}

	configureSlackUserGroupChannels(d, userGroup)

	return nil
}

func resourceSlackUserGroupChannelsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	currentId := d.Id()
	usergroupId := d.Get("usergroup_id").(string)

	log.Printf("[DEBUG] Deleting usergroup channels relation: %s", usergroupId)

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

	if _, err := client.UpdateUserGroupContext(ctx, *params); err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("provider cannot remove all default channels from the slack usergroup (%s)", usergroupId),
				Detail:   err.Error(),
			},
		}
	}

	d.SetId("")

	return nil
}
