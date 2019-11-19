package slack

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceGroup() *schema.Resource {
	return resourceSlackGroup()
}
