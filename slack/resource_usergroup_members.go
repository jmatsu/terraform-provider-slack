package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"log"
	"strings"
)

func resourceSlackUserGroupMembers() *schema.Resource {
	return &schema.Resource{
		Read:   resourceSlackUserGroupMembersRead,
		Create: resourceSlackUserGroupMembersCreate,
		Update: resourceSlackUserGroupMembersUpdate,
		Delete: resourceSlackUserGroupMembersDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				_ = d.Set("usergroup_id", d.Id())
				return schema.ImportStatePassthrough(d, m)
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

func resourceSlackUserGroupMembersCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

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
		return err
	}

	configureSlackUserGroupMembers(d, userGroup)

	return nil
}

func resourceSlackUserGroupMembersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	usergroupId := d.Get("usergroup_id").(string)

	log.Printf("[DEBUG] Reading usergroup members: %s", usergroupId)

	if usergroupId != d.Id() {
		return fmt.Errorf("it looks usergroup id has been changed but it's not allowed. Res ID: %s", d.Id())
	}

	members, err := client.GetUserGroupMembersContext(ctx, usergroupId)

	if err != nil {
		return err
	}

	_ = d.Set("members", members)

	return nil
}

func resourceSlackUserGroupMembersUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	usergroupId := d.Get("usergroup_id").(string)

	if usergroupId != d.Id() {
		return fmt.Errorf("it looks usergroup id has been changed but it's not allowed. Res ID: %s", d.Id())
	}

	_, err := client.EnableUserGroupContext(ctx, usergroupId)

	if err != nil && err.Error() != "already_enabled" {
		return err
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
		return err
	}

	configureSlackUserGroupMembers(d, userGroup)

	return nil
}

func resourceSlackUserGroupMembersDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	usergroupId := d.Get("usergroup_id").(string)

	if usergroupId != d.Id() {
		return fmt.Errorf("it looks usergroup id has been changed but it's not allowed. Res ID: %s", d.Id())
	}

	log.Printf("[DEBUG] Reading usergroup members: %s", usergroupId)

	// Cannot use "" as a member parameter, so let me disable it
	if _, err := client.DisableUserGroupContext(ctx, usergroupId); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
