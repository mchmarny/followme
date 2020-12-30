package format

// ToYesNo returns "yes" for true and "no" for false
func ToYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
