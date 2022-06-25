package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"log"
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

func configureSlackUserGroupMembers(d *schema.ResourceData, userGroup slack.UserGroup) {
	d.SetId(userGroup.ID)
	_ = d.Set("members", userGroup.Users)
}

func resourceSlackUserGroupMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	usergroupId := d.Get("usergroup_id").(string)

	iMembers := d.Get("members").(*schema.Set)
	userIds := make([]string, len(iMembers.List()))
	for i, v := range iMembers.List() {
		userIds[i] = v.(string)
	}
	userIdParam := strings.Join(userIds, ",")

	log.Printf("[DEBUG] Creating usergroup members: %s (%s)", usergroupId, userIdParam)

	userGroup, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, userIdParam)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't attach members of the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   "https://api.slack.com/methods/usergroups.users.update",
			},
		}
	}

	configureSlackUserGroupMembers(d, userGroup)

	return nil
}

func resourceSlackUserGroupMembersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	currentId := d.Id()
	usergroupId := d.Get("usergroup_id").(string)

	log.Printf("[DEBUG] Reading usergroup members: %s", usergroupId)

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
				Detail:   "https://api.slack.com/methods/usergroups.users.list",
			},
		}
	}

	_ = d.Set("members", members)

	return nil
}

func resourceSlackUserGroupMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	currentId := d.Id()
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

	_, err := client.EnableUserGroupContext(ctx, usergroupId)

	if err != nil && err.Error() != "already_enabled" {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't activate the slack usergroup (%s) to update members due to *%s*", usergroupId, err.Error()),
				Detail:   "https://api.slack.com/methods/usergroups.users.enable",
			},
		}
	}

	iMembers := d.Get("members").(*schema.Set)
	userIds := make([]string, len(iMembers.List()))
	for i, v := range iMembers.List() {
		userIds[i] = v.(string)
	}
	userIdParam := strings.Join(userIds, ",")

	log.Printf("[DEBUG] Updating usergroup members: %s (%s)", usergroupId, userIdParam)

	userGroup, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, userIdParam)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't update members of the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   "https://api.slack.com/methods/usergroups.users.update",
			},
		}
	}

	configureSlackUserGroupMembers(d, userGroup)

	return nil
}

func resourceSlackUserGroupMembersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client

	currentId := d.Id()
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

	log.Printf("[DEBUG] Reading usergroup members: %s", usergroupId)

	// Cannot use "" as a member parameter, so let me disable it
	if _, err := client.DisableUserGroupContext(ctx, usergroupId); err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't disable the slack usergroup (%s) due to *%s*", usergroupId, err.Error()),
				Detail:   "https://api.slack.com/methods/usergroups.disable",
			},
		}
	}

	d.SetId("")

	return nil
}
