package rafay

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// copySchemaMap returns a shallow copy of the provided schema map so callers
// can safely mutate it without touching shared state.
func copySchemaMap(src map[string]*schema.Schema) map[string]*schema.Schema {
	if src == nil {
		return nil
	}
	dst := make(map[string]*schema.Schema, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
