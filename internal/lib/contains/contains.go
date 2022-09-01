package contains

// String returns true if a string is present in an iteratee.
func String(s []string, v string) bool {
	for _, val := range s {
		if val == v {
			return true
		}
	}
	return false
}
