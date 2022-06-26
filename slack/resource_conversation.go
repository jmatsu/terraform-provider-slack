package slack

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/slack-go/slack"
)

const (
	conversationActionOnDestroyNone    = "none"
	conversationActionOnDestroyArchive = "archive"
)

var validateConversationActionOnDestroyValue = validation.StringInSlice([]string{
	conversationActionOnDestroyNone,
	conversationActionOnDestroyArchive,
}, false)

func resourceSlackConversation() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSlackConversationRead,
		CreateContext: resourceSlackConversationCreate,
		UpdateContext: resourceSlackConversationUpdate,
		DeleteContext: resourceSlackConversationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Required: true,
			},
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
			"action_on_destroy": {
				Type:         schema.TypeString,
				Description:  "Either of none or archive",
				Required:     true,
				ValidateFunc: validateConversationActionOnDestroyValue,
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

func configureSlackConversation(ctx context.Context, logger *Logger, d *schema.ResourceData, channel *slack.Channel) {
	d.SetId(channel.ID)
	_ = d.Set("name", channel.Name)
	_ = d.Set("topic", channel.Topic.Value)
	_ = d.Set("purpose", channel.Purpose.Value)
	_ = d.Set("is_archived", channel.IsArchived)
	_ = d.Set("is_shared", channel.IsShared)
	_ = d.Set("is_ext_shared", channel.IsExtShared)
	_ = d.Set("is_org_shared", channel.IsOrgShared)
	_ = d.Set("created", channel.Created)
	_ = d.Set("creator", channel.Creator)

	// Required
	_ = d.Set("is_private", channel.IsPrivate)

	// Never support
	//_ = d.Set("members", channel.Members)
	//_ = d.Set("num_members", channel.NumMembers)
	//_ = d.Set("unread_count", channel.UnreadCount)
	//_ = d.Set("unread_count_display", channel.UnreadCountDisplay)
	//_ = d.Set("last_read", channel.Name)
	//_ = d.Set("latest", channel.Name)

	logger.debug(ctx, "Configured Conversation #%s (isArchived = %t)", d.Id(), d.Get("is_archived").(bool))
}

func resourceSlackConversationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	isPrivate := d.Get("is_private").(bool)

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":          "slack_conversation",
		"conversation_name": name,
		"is_private":        isPrivate,
	})

	logger.trace(ctx, "Start creating a conversation")

	channel, err := client.CreateConversationContext(ctx, name, isPrivate)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't create a slack conversation (%s, isPrivate = %t) due to *%s*", name, isPrivate, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.create"),
			},
		}
	} else {
		logger.trace(ctx, "Got a response from Slack API")
	}

	configureSlackConversation(ctx, logger, d, channel)

	return nil
}

func resourceSlackConversationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":        "slack_conversation",
		"conversation_id": id,
	})

	logger.trace(ctx, "Start reading the conversation")

	channel, err := client.GetConversationInfoContext(ctx, id, false)

	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't find a slack conversation (%s) due to *%s*", id, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.info"),
			},
		}
	} else {
		logger.trace(ctx, "Got a response from Slack API")
	}

	configureSlackConversation(ctx, logger, d, channel)

	return nil
}

func resourceSlackConversationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":        "slack_conversation",
		"conversation_id": id,
	})

	// TODO check if it's changed or not to reduce api calls

	name := d.Get("name").(string)

	if _, err := client.RenameConversationContext(ctx, id, name); err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Slack provider couldn't rename a slack conversation (%s) to %s due to *%s*", id, name, err.Error()),
				Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.rename"),
			},
		}
	} else {
		logger.trace(ctx, "Renamed the conversation to %s", name)
	}

	if topic, ok := d.GetOk("topic"); ok {
		if _, err := client.SetTopicOfConversationContext(ctx, id, topic.(string)); err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Slack provider couldn't set a topic of a slack conversation (%s) to %s due to *%s*", id, topic.(string), err.Error()),
					Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.setTopic"),
				},
			}
		}
	} else {
		logger.trace(ctx, "Set the conversation topic to %s", topic)
	}

	if purpose, ok := d.GetOk("purpose"); ok {
		if _, err := client.SetPurposeOfConversationContext(ctx, id, purpose.(string)); err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Slack provider couldn't set a purpose of a slack conversation (%s) to %s due to *%s*", id, purpose.(string), err.Error()),
					Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.setPurpose"),
				},
			}
		}
	}

	if isArchived, ok := d.GetOkExists("is_archived"); ok {
		if isArchived.(bool) {
			if err := client.ArchiveConversationContext(ctx, id); err != nil {
				if err.Error() != "already_archived" {
					return diag.Diagnostics{
						{
							Severity: diag.Error,
							Summary:  fmt.Sprintf("Slack provider couldn't archive a slack conversation (%s) due to *%s*", id, err.Error()),
							Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.archive"),
						},
					}
				} else {
					logger.debug(ctx, "The conversation has already been archived")
				}
			}

			logger.trace(ctx, "Archived the conversation")
		} else {
			if err := client.UnArchiveConversationContext(ctx, id); err != nil {
				if err.Error() != "not_archived" {
					return diag.Diagnostics{
						{
							Severity: diag.Error,
							Summary:  fmt.Sprintf("Slack provider couldn't unarchive a slack conversation (%s) due to *%s*", id, err.Error()),
							Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.unarchive"),
						},
					}
				} else {
					logger.debug(ctx, "The conversation has already been unarchived")
				}
			}

			logger.trace(ctx, "Unarchived the conversation")
		}
	}

	return resourceSlackConversationRead(ctx, d, meta)
}

func resourceSlackConversationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	client := meta.(*Team).client
	logger := meta.(*Team).logger.withTags(map[string]interface{}{
		"resource":        "slack_conversation",
		"conversation_id": id,
	})

	action := d.Get("action_on_destroy").(string)

	switch action {
	case conversationActionOnDestroyNone:
		logger.debug(ctx, "Does nothing on destroy")
	case conversationActionOnDestroyArchive:
		logger.debug(ctx, "Archive the conversation (%s) on destroy", d.Get("name").(string))

		if err := client.ArchiveConversationContext(ctx, id); err != nil {
			if err.Error() != "already_archived" {
				return diag.Diagnostics{
					{
						Severity: diag.Error,
						Summary:  fmt.Sprintf("Slack provider couldn't archive a slack conversation (%s) due to *%s*", id, err.Error()),
						Detail:   fmt.Sprintf("Please refer to %s for the details.", "https://api.slack.com/methods/conversations.archive"),
					},
				}
			} else {
				logger.debug(ctx, "The conversation has already been archived")
			}
		}

		logger.trace(ctx, "Archived the conversation")
	default:
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("%s in action_on_destroy is not acceptable", action),
				Detail:   fmt.Sprintf("Either one of %s and %s is allowed", conversationActionOnDestroyNone, conversationActionOnDestroyArchive),
			},
		}
	}

	d.SetId("")

	logger.debug(ctx, "Cleared the resource id of this conversation so it's going to be removed from the state")

	return nil
}
