package slack

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/nlopes/slack"
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

func configureSlackGroup(d *schema.ResourceData, group *slack.Group) {
	d.SetId(group.ID)
	_ = d.Set("name", group.Name)
	_ = d.Set("topic", group.Topic.Value)
	_ = d.Set("purpose", group.Purpose.Value)
	_ = d.Set("is_archived", group.IsArchived)
	_ = d.Set("is_shared", group.IsShared)
	_ = d.Set("is_ext_shared", group.IsExtShared)
	_ = d.Set("is_org_shared", group.IsOrgShared)
	_ = d.Set("created", group.Created)
	_ = d.Set("creator", group.Creator)

	// Never support
	//_ = d.Set("members", Group.Members)
	//_ = d.Set("num_members", Group.NumMembers)
	//_ = d.Set("unread_count", Group.UnreadCount)
	//_ = d.Set("unread_count_display", Group.UnreadCountDisplay)
	//_ = d.Set("last_read", Group.Name)
	//_ = d.Set("latest", Group.Name)
}

func resourceSlackGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	name := d.Get("name").(string)

	newGroup := name

	ctx := context.Background()

	log.Printf("[DEBUG] Creating Group: %s", name)
	group, err := client.CreateGroupContext(ctx, newGroup)

	if err != nil {
		return err
	}

	configureSlackGroup(d, group)

	return err
}

func resourceSlackGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Reading Group: %s", d.Id())
	group, err := client.GetGroupInfoContext(ctx, id)

	if err != nil {
		return err
	}

	configureSlackGroup(d, group)

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
			if err := client.ArchiveGroupContext(ctx, id); err != nil {
				if err.Error() != "already_archived" {
					return err
				}
			}
		} else {
			if err := client.UnarchiveGroupContext(ctx, id); err != nil {
				if err.Error() != "not_archived" {
					return err
				}
			}
		}
	}

	return resourceSlackGroupRead(d, meta)
}

func resourceSlackGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	id := d.Id()

	log.Printf("[DEBUG] Deleting(archive) Group: %s (%s)", id, d.Get("name"))

	if err := client.ArchiveGroupContext(ctx, id); err != nil {
		return err
	}

	d.SetId("")
	return nil
}
