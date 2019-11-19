package slack

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceSlackChannel() *schema.Resource {
	return &schema.Resource{
		Read:   resourceSlackChannelRead,
		Create: resourceSlackChannelCreate,
		Update: resourceSlackChannelUpdate,
		Delete: resourceSlackChannelDelete,

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
			"locale": {
				Type:     schema.TypeString,
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

func resourceSlackChannelCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	name := d.Get("name").(string)

	newChannel := name

	ctx := context.Background()

	log.Printf("[DEBUG] Creating Channel: %s", name)
	_, err := client.CreateChannelContext(ctx, newChannel)

	return err
}

func resourceSlackChannelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Reading Channel: %s", d.Id())
	channel, err := client.GetChannelInfoContext(ctx, id)

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
	_ = d.Set("name", channel.Name)
	_ = d.Set("topic", channel.Topic.Value)
	_ = d.Set("purpose", channel.Purpose.Value)
	_ = d.Set("is_archived", channel.IsArchived)
	_ = d.Set("is_shared", channel.IsShared)
	_ = d.Set("is_ext_shared", channel.IsExtShared)
	_ = d.Set("is_org_shared", channel.IsOrgShared)
	_ = d.Set("locale", channel.Locale)
	_ = d.Set("created", channel.Created)
	_ = d.Set("creator", channel.Creator)

	// Never support
	//_ = d.Set("members", channel.Members)
	//_ = d.Set("num_members", channel.NumMembers)
	//_ = d.Set("unread_count", channel.UnreadCount)
	//_ = d.Set("unread_count_display", channel.UnreadCountDisplay)
	//_ = d.Set("last_read", channel.Name)
	//_ = d.Set("latest", channel.Name)

	return nil
}

func resourceSlackChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	if _, err := client.RenameChannelContext(ctx, id, d.Get("name").(string)); err != nil {
		return err
	}

	if topic, ok := d.GetOk("topic"); ok {
		if _, err := client.SetChannelTopicContext(ctx, id, topic.(string)); err != nil {
			return err
		}
	}

	if purpose, ok := d.GetOk("purpose"); ok {
		if _, err := client.SetChannelPurposeContext(ctx, id, purpose.(string)); err != nil {
			return err
		}
	}

	if isArchived, ok := d.GetOkExists("is_archived"); ok {
		if isArchived.(bool) {
			if err := client.ArchiveGroupContext(ctx, id); err != nil {
				return err
			}
		} else {
			if err := client.UnarchiveGroupContext(ctx, id); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceSlackChannelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Deleting(archive) Channel: %s (%s)", id, d.Get("name"))

	if isArchived, ok := d.GetOkExists("is_archived"); ok && isArchived.(bool) {
		log.Printf("[DEBUG] Did nothing because this channel has already been archived. %s", id)
		return nil
	}

	return client.ArchiveChannelContext(ctx, id)
}
