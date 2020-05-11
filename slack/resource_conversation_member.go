package slack

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/slack-go/slack"
	"github.com/thedevsaddam/retry"
)

const slackConversationMemberErrAlreadyInChannel = "already_in_channel"
const slackConversationMemberErrNotInChannel = "not_in_channel"
const slackConversationMemberRetryAttempts = 3
const slackConversationMemberRetryDelay = 30 * time.Second

func resourceSlackConversationMember() *schema.Resource {
	return &schema.Resource{
		Read:   resourceSlackConversationMemberRead,
		Create: resourceSlackConversationMemberCreate,
		Delete: resourceSlackConversationMemberDelete,

		Schema: map[string]*schema.Schema{
			"conversation_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func configureSlackConversationMember(d *schema.ResourceData, conversationID string, userID string) {
	if conversationID != "" && userID != "" {
		d.SetId(fmt.Sprintf("%s-%s", conversationID, userID))
	}
}

func resourceSlackConversationMemberCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	conversationID := d.Get("conversation_id").(string)
	userID := d.Get("user_id").(string)

	err := retry.DoFunc(slackConversationMemberRetryAttempts, slackConversationMemberRetryDelay, func() error {
		log.Printf("[DEBUG] Inviting conversation member: %s %s", conversationID, userID)
		_, err := client.InviteUsersToConversationContext(ctx, conversationID, userID)
		if err != nil {
			if strings.Contains(err.Error(), slackConversationMemberErrAlreadyInChannel) {
				// user is already in channel. do not fail, consider it as a successful end state.
				return nil
			}
		}
		return err
	})
	if err != nil {
		return err
	}

	configureSlackConversationMember(d, conversationID, userID)
	return nil
}

func resourceSlackConversationMemberRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	conversationID := d.Get("conversation_id").(string)
	userID := d.Get("user_id").(string)

	log.Printf("[DEBUG] Reading conversation member: %s %s", conversationID, userID)
	memberIDs, _, err := client.GetUsersInConversationContext(ctx, &slack.GetUsersInConversationParameters{
		ChannelID: conversationID,
	})

	if err != nil {
		return err
	}

	for _, memberID := range memberIDs {
		if memberID == userID {
			configureSlackConversationMember(d, conversationID, userID)
			break
		}
	}

	return nil
}

func resourceSlackConversationMemberDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	conversationID := d.Get("conversation_id").(string)
	userID := d.Get("user_id").(string)

	err := retry.DoFunc(slackConversationMemberRetryAttempts, slackConversationMemberRetryDelay, func() error {
		log.Printf("[DEBUG] Deleting conversation member: %s %s", conversationID, userID)
		err := client.KickUserFromConversationContext(ctx, conversationID, userID)
		if err != nil {
			if strings.Contains(err.Error(), slackConversationMemberErrNotInChannel) {
				// user is already not in channel. do not fail, consider it as a successful end state.
				return nil
			}
		}
		return err
	})
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
