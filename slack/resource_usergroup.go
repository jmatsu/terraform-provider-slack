package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

const userGroupListCacheFileName = "usergroups.json"

func resourceSlackUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSlackUserGroupRead,
		CreateContext: resourceSlackUserGroupCreate,
		UpdateContext: resourceSlackUserGroupUpdate,
		DeleteContext: resourceSlackUserGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "",
				ValidateDiagFunc: validateEnums([]string{"admins", "owners", ""}),
			},
			"team_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func configureSlackUserGroup(ctx context.Context, logger *Logger, d *schema.ResourceData, userGroup slack.UserGroup) {
	d.SetId(userGroup.ID)
	_ = d.Set("handle", userGroup.Handle)
	_ = d.Set("name", userGroup.Name)
	_ = d.Set("description", userGroup.Description)
	_ = d.Set("auto_type", userGroup.AutoType)
	_ = d.Set("team_id", userGroup.TeamID)

	logger.debug(ctx, "Configured UserGroup #%s @%s", d.Id(), d.Get("handle").(string))
}

func resourceSlackUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	handle := d.Get("handle").(string)

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":         "slack_conversation",
		"usergroup_handle": handle,
	})

	logger.debug(ctx, "Start creating a usergroup")

	var name = handle

	if value, ok := d.GetOk("name"); ok {
		logger.debug(ctx, "usergroup name (%s) is specified", value.(string))
		name = value.(string)
	}

	newUserGroup := &slack.UserGroup{
		Handle:      handle,
		Name:        name,
		Description: d.Get("description").(string),
		AutoType:    d.Get("auto_type").(string),
	}

	userGroup, err := client.CreateUserGroupContext(ctx, *newUserGroup)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't create a slack usergroup (%s) due to *%s*", handle, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.create"),
			},
		}
	} else {
		logger.trace(ctx, "Got a response from Slack API")
	}

	configureSlackUserGroup(ctx, logger, d, userGroup)

	return nil
}

func resourceSlackUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_conversation",
		"usergroup_id": id,
	})

	logger.trace(ctx, "Start reading a usergroup")

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
					Summary:  fmt.Sprintf("Slack provider couldn't find slack usergroups due to *%s*", err.Error()),
					Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.list"),
				},
			}
		} else {
			logger.trace(ctx, "Got a response from Slack API")
		}

		userGroups = &tempUserGroups

		saveCacheAsJson(userGroupListCacheFileName, &userGroups)
	} else {
		logger.trace(ctx, "Read usergroups from the cahed")
	}

	for _, userGroup := range *userGroups {
		if userGroup.ID == id {
			configureSlackUserGroup(ctx, logger, d, userGroup)
			return nil
		}
	}

	return diag.Diagnostics{
		{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Slack provider couldn't find a slack usergroup (%s)", id),
			Detail:   fmt.Sprintf("a usergroup (%s) is not found in this workspace", id),
		},
	}
}

func resourceSlackUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_conversation",
		"usergroup_id": id,
	})

	logger.trace(ctx, "Start updating the usergroup")

	handle := d.Get("handle").(string)
	var name = handle

	if value, ok := d.GetOk("name"); ok {
		logger.debug(ctx, "name (%s) is specified", value.(string))
		name = value.(string)
	}

	editedUserGroup := &slack.UserGroup{
		ID:          id,
		Handle:      handle,
		Name:        name,
		Description: d.Get("description").(string),
		AutoType:    d.Get("auto_type").(string),
	}

	userGroup, err := client.UpdateUserGroupContext(ctx, *editedUserGroup)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't update the slack usergroup (%s) due to *%s*", id, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.update"),
			},
		}
	} else {
		logger.trace(ctx, "Got a response from Slack API")
	}

	configureSlackUserGroup(ctx, logger, d, userGroup)
	return nil
}

func resourceSlackUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_conversation",
		"usergroup_id": id,
	})

	logger.trace(ctx, "Start deleting (actually disabling) usergroup")

	if _, err := client.DisableUserGroupContext(ctx, id); err != nil {
		if err.Error() != "already_disabled" {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Slack provider couldn't disable the slack usergroup (%s) due to *%s*", id, err.Error()),
					Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.disable"),
				},
			}
		} else {
			logger.debug(ctx, "This usergroup has already been disabled")
		}
	} else {
		logger.trace(ctx, "Got a response from Slack API")
	}

	d.SetId("")

	logger.debug(ctx, "Cleared the resource id of this usergroup so it's going to be removed from the state")

	return nil
}
