package slack

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

func validateEnums(values []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		if !containsAny(values, v.(string)) {
			errors = append(errors, fmt.Errorf("%s is an invalid value for argument %s", v.(string), k))
		}
		return
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
