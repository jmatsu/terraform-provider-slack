package slack

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/slack-go/slack"
)

func dataSourceConversation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSlackConversationRead,

		Schema: map[string]*schema.Schema{
			"channel_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"channel_name": {
				Type:     schema.TypeString,
				Optional: true,
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

func dataSlackConversationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Team).client
	conversationId := d.Get("channel_id").(string)
	conversationName := d.Get("channel_name").(string)

	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"data":            "conversation",
		"conversation_id": conversationId,
	})

	logger.trace(ctx, "Start reading a conversation")

	var channel *slack.Channel
	var err error
	if conversationName != "" {
		channel, err = findConversationByName(client, conversationName)
	} else {
		channel, err = client.GetConversationInfoContext(ctx, conversationId, false)
	}

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't read conversation %s due to *%s*", conversationId, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.info"),
			},
		}
	} else {
		logger.trace(ctx, "Got a response from Slack api")
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

	logger.debug(ctx, "Conversation #%s (isArchived = %t)", d.Get("name").(string), d.Get("is_archived").(bool))

	return nil
}

func findConversationByName(client *slack.Client, name string) (*slack.Channel, error) {
	channels, _, err := client.GetConversations(&slack.GetConversationsParameters{Limit: 1024})
	if err != nil {
		return nil, err
	}

	for _, c := range channels {
		if c.Name == name {
			return &c, nil
		}
	}

	return nil, errors.New("channel not found")
}
