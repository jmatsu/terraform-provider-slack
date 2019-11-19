package slack

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nlopes/slack"
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

func configureSlackChannel(d *schema.ResourceData, channel *slack.Channel) {
	d.SetId(channel.ID)
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
}

func resourceSlackChannelCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	name := d.Get("name").(string)

	newChannel := name

	ctx := context.Background()

	log.Printf("[DEBUG] Creating Channel: %s", name)
	channel, err := client.CreateChannelContext(ctx, newChannel)

	if err != nil {
		return err
	}

	configureSlackChannel(d, channel)

	return nil
}

func resourceSlackChannelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Reading Channel: %s", d.Id())
	channel, err := client.GetChannelInfoContext(ctx, id)

	if err != nil {
		return err
	}

	configureSlackChannel(d, channel)

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
			if err := client.ArchiveChannelContext(ctx, id); err != nil {
				if err.Error() != "already_archived" {
					return err
				}
			}
		} else {
			if err := client.UnarchiveChannelContext(ctx, id); err != nil {
				if err.Error() != "not_archived" {
					return err
				}
			}
		}
	}

	return resourceSlackChannelRead(d, meta)
}

func resourceSlackChannelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Deleting(archive) Channel: %s (%s)", id, d.Get("name"))

	if err := client.ArchiveChannelContext(ctx, id); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
