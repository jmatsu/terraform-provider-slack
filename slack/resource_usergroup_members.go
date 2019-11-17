package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
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
			State: schema.ImportStatePassthrough,
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

func resourceSlackUserGroupMembersCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceSlackUserGroupMembersUpdate(d, meta)
}

func resourceSlackUserGroupMembersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	usergroupId := d.Get("usergroup_id").(string)

	log.Printf("[DEBUG] Reading usergroup members: %s", usergroupId)

	members, err := client.GetUserGroupMembersContext(ctx, usergroupId)

	if err != nil {
		return err
	}

	_ = d.Set("usergroup_id", usergroupId)
	_ = d.Set("members", members)

	return nil
}

func resourceSlackUserGroupMembersUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	usergroupId := d.Get("usergroup_id").(string)
	iMembers := d.Get("members").([]interface{})
	userIds := make([]string, len(iMembers))
	for i, v := range iMembers {
		userIds[i] = v.(string)
	}
	userIdParam := strings.Join(userIds, ",")

	log.Printf("[DEBUG] Updating usergroup members: %s (%s)", usergroupId, userIdParam)

	newUserGroup, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, userIdParam)

	if err != nil {
		return err
	}

	d.SetId(resourceSlackUserGroupMembersId(newUserGroup.ID, newUserGroup.Handle))
	return resourceSlackUserGroupMembersRead(d, meta)
}

func resourceSlackUserGroupMembersDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	usergroupId := d.Get("usergroup_id").(string)

	log.Printf("[DEBUG] Reading usergroup members: %s", usergroupId)

	_, err := client.UpdateUserGroupMembersContext(ctx, usergroupId, "")
	return err
}

func resourceSlackUserGroupMembersUserGroupId(d *schema.ResourceData) string {
	return strings.Split(d.Id(), ":")[0]
}

func resourceSlackUserGroupMembersId(usergroupId string, handle string) string {
	return fmt.Sprintf("%s:%s", usergroupId, handle)
}
