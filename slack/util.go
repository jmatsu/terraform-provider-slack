package slack

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func validateEnums(values []string) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		if !containsAny(values, v.(string)) {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("%s is an invalid value for argument %s", v.(string), path),
					Detail:   "",
				},
			}
		}

		return nil
	}
}

func containsAny(values []string, any string) bool {
	valid := false

	for _, value := range values {
		if value == any {
			valid = true
			break
		}
	}

	return valid
}
