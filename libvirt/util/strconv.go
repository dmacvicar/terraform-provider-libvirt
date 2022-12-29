package util

// FormatBoolYesNo is similar to strconv.FormatBool with yes/no instead of true/false.
func FormatBoolYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
