package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
	"log"
)

const userGroupListCacheFileName = "usergroups.json"

func resourceSlackUserGroup() *schema.Resource {
	return &schema.Resource{
		Read:   resourceSlackUserGroupRead,
		Create: resourceSlackUserGroupCreate,
		Update: resourceSlackUserGroupUpdate,
		Delete: resourceSlackUserGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateEnums([]string{"admins", "owners", ""}),
			},
			"team_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func configureSlackUserGroup(d *schema.ResourceData, userGroup slack.UserGroup) {
	d.SetId(userGroup.ID)
	_ = d.Set("handle", userGroup.Handle)
	_ = d.Set("name", userGroup.Name)
	_ = d.Set("description", userGroup.Description)
	_ = d.Set("auto_type", userGroup.AutoType)
	_ = d.Set("team_id", userGroup.TeamID)
}

func resourceSlackUserGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	handle := d.Get("handle").(string)
	var name = handle

	if _, ok := d.GetOk("name"); ok {
		name = d.Get("name").(string)
	}

	newUserGroup := &slack.UserGroup{
		Handle:      handle,
		Name:        name,
		Description: d.Get("description").(string),
		AutoType:    d.Get("auto_type").(string),
	}

	ctx := context.Background()

	log.Printf("[DEBUG] Creating usergroup: %s (%s)", handle, name)
	userGroup, err := client.CreateUserGroupContext(ctx, *newUserGroup)

	if err != nil {
		return err
	}

	configureSlackUserGroup(d, userGroup)

	return nil
}

func resourceSlackUserGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Reading usergroup: %s", d.Id())

	// Use a cache for usergroups api call because the limitation is strict
	var userGroups *[]slack.UserGroup

	if !restoreJsonCache(userGroupListCacheFileName, &userGroups) {
		tempUserGroups, err := client.GetUserGroupsContext(ctx, func(params *slack.GetUserGroupsParams) {
			params.IncludeUsers = false
			params.IncludeCount = false
			params.IncludeDisabled = true
		})

		if err != nil {
			return err
		}

		userGroups = &tempUserGroups

		saveCacheAsJson(userGroupListCacheFileName, &userGroups)
	}

	for _, userGroup := range *userGroups {
		if userGroup.ID == id {
			configureSlackUserGroup(d, userGroup)
			return nil
		}
	}

	return fmt.Errorf("a usergroup (%s) is not found", id)
}

func resourceSlackUserGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	handle := d.Get("handle").(string)
	var name = handle

	if _, ok := d.GetOk("name"); ok {
		name = d.Get("name").(string)
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	editedUserGroup := &slack.UserGroup{
		ID:          id,
		Handle:      handle,
		Name:        name,
		Description: d.Get("description").(string),
		AutoType:    d.Get("auto_type").(string),
	}

	log.Printf("[DEBUG] Updating usergroup: %s", d.Id())
	userGroup, err := client.UpdateUserGroupContext(ctx, *editedUserGroup)

	if err != nil {
		return err
	}

	configureSlackUserGroup(d, userGroup)
	return nil
}

func resourceSlackUserGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Deleting usergroup: %s", id)
	if _, err := client.DisableUserGroupContext(ctx, id); err != nil {
		if err.Error() != "already_disabled" {
			return err
		}
	}

	d.SetId("")

	return nil
}
