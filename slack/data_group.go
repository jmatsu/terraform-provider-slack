package slack

import (
	"context"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func dataSourceGroup() *schema.Resource {

	return &schema.Resource{
		Read: dataSlackGroupRead,

		Schema: map[string]*schema.Schema{
			"group_id": {
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
			},
			"purpose": {
				Type:     schema.TypeString,
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

func dataSlackGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Team).client

	groupId := d.Get("group_id").(string)

	ctx := context.WithValue(context.Background(), ctxId, groupId)

	log.Printf("[DEBUG] Reading group: %s", groupId)
	group, err := client.GetGroupInfoContext(ctx, groupId)

	if err != nil {
		return err
	}

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

	return nil
}
