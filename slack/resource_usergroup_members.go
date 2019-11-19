package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nlopes/slack"
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
				Type: schema.TypeList,
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
	d.SetId(d.Get("usergroup_id").(string))
	return resourceSlackUserGroupMembersUpdate(d, meta)
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

	iMembers := d.Get("members").([]interface{})
	userIds := make([]string, len(iMembers))
	for i, v := range iMembers {
		userIds[i] = v.(string)
	}
	userIdParam := strings.Join(userIds, ",")

	log.Printf("[DEBUG] Updating usergroup members: %s (%s)", usergroupId, userIdParam)

	userGroup, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, userIdParam)

	if err != nil {
		return err
	}

	configureSlackUserGroupMembers(d, userGroup)

	return resourceSlackUserGroupMembersRead(d, meta)
}

func resourceSlackUserGroupMembersDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	usergroupId := d.Get("usergroup_id").(string)

	if usergroupId != d.Id() {
		return fmt.Errorf("it looks usergroup id has been changed but it's not allowed. Res ID: %s", d.Id())
	}

	log.Printf("[DEBUG] Reading usergroup members: %s", usergroupId)

	if _, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, ""); err != nil {
		return err
	}

	d.SetId("")
	return nil
}
