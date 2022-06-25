package slack

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceConversation() *schema.Resource {
	return &schema.Resource{
		Read: dataSlackConversationRead,

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"topic": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"purpose": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_archived": {
				Type:     schema.TypeBool,
				Computed: true,
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

func dataSlackConversationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	channelId := d.Get("channel_id").(string)

	ctx := context.WithValue(context.Background(), ctxId, channelId)

	log.Printf("[DEBUG] Reading Conversation: %s", channelId)
	channel, err := client.GetConversationInfoContext(ctx, channelId, false)

	if err != nil {
		return err
	}

	d.SetId(channel.ID)
	_ = d.Set("name", channel.Name)
	_ = d.Set("topic", channel.Topic.Value)
	_ = d.Set("purpose", channel.Purpose.Value)
	_ = d.Set("is_private", channel.IsPrivate)
	_ = d.Set("is_archived", channel.IsArchived)
	_ = d.Set("is_shared", channel.IsShared)
	_ = d.Set("is_ext_shared", channel.IsExtShared)
	_ = d.Set("is_org_shared", channel.IsOrgShared)
	_ = d.Set("created", channel.Created)
	_ = d.Set("creator", channel.Creator)

	return nil
}
