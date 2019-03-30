package slack

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nlopes/slack"
	"log"
)

func resourceSlackUserGroup() *schema.Resource {
	return &schema.Resource{
		Read:   resourceSlackUserGroupRead,
		Create: resourceSlackuserGroupCreate,
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
				Type:     schema.TypeInt,
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
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"team_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	ctxId = 1
)

func resourceSlackuserGroupCreate(d *schema.ResourceData, meta interface{}) error {
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

	d.SetId(userGroup.ID)
	return resourceSlackUserGroupRead(d, meta)
}

func resourceSlackUserGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Reading usergroup: %s", d.Id())
	groups, err := client.GetUserGroupsContext(ctx, func(params *slack.GetUserGroupsParams) {
		params.IncludeUsers = false
		params.IncludeCount = false
		params.IncludeDisabled = false
	})

	if err != nil {
		switch err.Error() {
		case "is_bot":
			log.Printf("[ERROR] Cannot call this api because a token is of bot")
			break
		case "missing_scope":
			log.Printf("[ERROR] Cannot call this api because a token does not have enough scope")
			break
		case "account_inactive":
			log.Printf("[ERROR] Cannot call this api because a token is of deleted user")
			break
		case "invalid_auth":
			log.Printf("[ERROR] Cannot call this api because a token is invalid or filtered by ip whitelist")
			break

		}

		return err
	}

	for _, group := range groups {
		if group.ID == id {
			_ = d.Set("handle", group.Handle)
			_ = d.Set("name", group.Name)
			_ = d.Set("description", group.Description)
			_ = d.Set("auto_type", group.AutoType)
			_ = d.Set("team_id", group.TeamID)
			return nil
		}
	}

	log.Printf("[WARN] Removing usergroup %s from state because it no longer exists in Slack",
		id)
	d.SetId("")

	return nil
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

	d.SetId(userGroup.ID)
	return resourceSlackUserGroupRead(d, meta)
}

func resourceSlackUserGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Deleting usergroup: %s", id)
	_, err := client.DisableUserGroupContext(ctx, id)
	return err
}
