package slack

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceChannel() *schema.Resource {
	return resourceSlackChannel()
}
