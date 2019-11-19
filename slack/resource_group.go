package slack

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceSlackGroup() *schema.Resource {
	return &schema.Resource{
		Read:   resourceSlackGroupRead,
		Create: resourceSlackGroupCreate,
		Update: resourceSlackGroupUpdate,
		Delete: resourceSlackGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// TODO support but api client's matter
			//"validate": {
			//	Type: schema.TypeBool,
			//	Optional:true,
			//	Default:true,
			//},
			"topic": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"purpose": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_archived": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"is_shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_ext_shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_org_shared": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"creator": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSlackGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	name := d.Get("name").(string)

	newGroup := name

	ctx := context.Background()

	log.Printf("[DEBUG] Creating Group: %s", name)
	_, err := client.CreateGroupContext(ctx, newGroup)

	return err
}

func resourceSlackGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Reading Group: %s", d.Id())
	Group, err := client.GetGroupInfoContext(ctx, id)

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

	_ = d.Set("name", Group.Name)
	_ = d.Set("topic", Group.Topic.Value)
	_ = d.Set("purpose", Group.Purpose.Value)
	_ = d.Set("is_archived", Group.IsArchived)
	_ = d.Set("is_shared", Group.IsShared)
	_ = d.Set("is_ext_shared", Group.IsExtShared)
	_ = d.Set("is_org_shared", Group.IsOrgShared)
	_ = d.Set("created", Group.Created)
	_ = d.Set("creator", Group.Creator)

	// Never support
	//_ = d.Set("members", Group.Members)
	//_ = d.Set("num_members", Group.NumMembers)
	//_ = d.Set("unread_count", Group.UnreadCount)
	//_ = d.Set("unread_count_display", Group.UnreadCountDisplay)
	//_ = d.Set("last_read", Group.Name)
	//_ = d.Set("latest", Group.Name)

	return nil
}

func resourceSlackGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	if _, err := client.RenameGroupContext(ctx, id, d.Get("name").(string)); err != nil {
		return err
	}

	if topic, ok := d.GetOk("topic"); ok {
		if _, err := client.SetGroupTopicContext(ctx, id, topic.(string)); err != nil {
			return err
		}
	}

	if purpose, ok := d.GetOk("purpose"); ok {
		if _, err := client.SetGroupPurposeContext(ctx, id, purpose.(string)); err != nil {
			return err
		}
	}

	if isArchived, ok := d.GetOkExists("is_archived"); ok {
		if isArchived.(bool) {
			if err := client.ArchiveChannelContext(ctx, id); err != nil {
				return err
			}
		} else {
			if err := client.UnarchiveChannelContext(ctx, id); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceSlackGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Deleting(archive) Group: %s (%s)", id, d.Get("name"))

	if isArchived, ok := d.GetOkExists("is_archived"); ok && isArchived.(bool) {
		log.Printf("[DEBUG] Did nothing because this group has already been archived. %s", id)
		return nil
	}

	return client.ArchiveGroupContext(ctx, id)
}
