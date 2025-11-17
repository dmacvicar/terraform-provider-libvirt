package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// mustSingleNestedAttribute asserts that attr is a schema.SingleNestedAttribute and panics otherwise.
// Generated schemas are known at compile time, so a panic indicates a programmer error.
func mustSingleNestedAttribute(attr schema.Attribute, name string) schema.SingleNestedAttribute {
	nested, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		panic(fmt.Sprintf("%s schema attribute is not a schema.SingleNestedAttribute", name))
	}
	return nested
}
