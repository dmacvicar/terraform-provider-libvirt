package suppress

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func CaseDifference(_, old, new string, _ *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}
