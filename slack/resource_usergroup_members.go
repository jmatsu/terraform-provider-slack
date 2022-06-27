package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"strings"
)

func resourceSlackUserGroupMembers() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSlackUserGroupMembersRead,
		CreateContext: resourceSlackUserGroupMembersCreate,
		UpdateContext: resourceSlackUserGroupMembersUpdate,
		DeleteContext: resourceSlackUserGroupMembersDelete,

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
			"members": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func configureSlackUserGroupMembers(ctx context.Context, logger *Logger, d *schema.ResourceData, userGroup slack.UserGroup) {
	d.SetId(userGroup.ID)
	_ = d.Set("members", userGroup.Users)

	logger.debug(ctx, "Configured channel members")
}

func resourceSlackUserGroupMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	usergroupId := d.Get("usergroup_id").(string)

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": usergroupId,
	})

	logger.trace(ctx, "Start creating members of the usergroup")

	iMembers := d.Get("members").(*schema.Set)
	userIds := make([]string, len(iMembers.List()))
	for i, v := range iMembers.List() {
		userIds[i] = v.(string)
	}
	userIdParam := strings.Join(userIds, ",")

	userGroup, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, userIdParam)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't attach members of the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.users.update"),
			},
		}
	}

	configureSlackUserGroupMembers(ctx, logger, d, userGroup)

	return nil
}

func resourceSlackUserGroupMembersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	currentId := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": currentId,
	})

	logger.trace(ctx, "Start reading the usergroup members")

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

	members, err := client.GetUserGroupMembersContext(ctx, usergroupId)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't read members of the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.users.list"),
			},
		}
	}

	_ = d.Set("members", members)

	logger.debug(ctx, "Configured members of the usergroup")

	return nil
}

func resourceSlackUserGroupMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	currentId := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": currentId,
	})

	logger.trace(ctx, "Start updating members of the usergroup")

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

	logger.debug(ctx, "Enable the usergroup first because disabled usergroups reject updates")
	_, err := client.EnableUserGroupContext(ctx, usergroupId)

	if err != nil && err.Error() != "already_enabled" {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't activate the slack usergroup (%s) to update members due to *%s*", usergroupId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.enable"),
			},
		}
	}

	iMembers := d.Get("members").(*schema.Set)
	userIds := make([]string, len(iMembers.List()))
	for i, v := range iMembers.List() {
		userIds[i] = v.(string)
	}
	userIdParam := strings.Join(userIds, ",")

	userGroup, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, userIdParam)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't update members of the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.users.update"),
			},
		}
	}

	configureSlackUserGroupMembers(ctx, logger, d, userGroup)

	return nil
}

func resourceSlackUserGroupMembersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	currentId := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":     "slack_usergroup_channels",
		"usergroup_id": currentId,
	})

	logger.trace(ctx, "Start destroying members of the usergroup")

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

	logger.debug(ctx, "A usergroup that has no members cannot be created by web API so just disable it")

	// Cannot use "" as a member parameter, so let me disable it
	if _, err := client.DisableUserGroupContext(ctx, usergroupId); err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't disable the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/usergroups.disable"),
			},
		}
	}

	d.SetId("")

	logger.debug(ctx, "Cleared the resource id of this usergroup members' resource so it's going to be removed from the state")

	return nil
}
