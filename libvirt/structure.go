package libvirt

// terraform uses lists with MaxItems=1 to simulate groups.
// this function packs a map[string]interface{} into an array for such usage
func flattenAsArray(m map[string]interface{}) []interface{} {
	if m == nil {
		return []interface{}{}
	}
	return []interface{}{m}
}
