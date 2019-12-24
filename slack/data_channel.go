package slack

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func dataSourceChannel() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "please use conversation resource with is_private=false instead because slack has deprecated this resource and related APIs.",

		Read: dataSlackChannelRead,

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

func dataSlackChannelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	channelId := d.Get("channel_id").(string)

	ctx := context.WithValue(context.Background(), ctxId, channelId)

	log.Printf("[DEBUG] Reading Channel: %s", channelId)
	channel, err := client.GetChannelInfoContext(ctx, channelId)

	if err != nil {
		return err
	}

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

	return nil
}
