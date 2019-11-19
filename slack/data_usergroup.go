package slack

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceUserGroup() *schema.Resource {
	return resourceSlackUserGroup()
}
