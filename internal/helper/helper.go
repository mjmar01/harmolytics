package helper

func StringInSlice(s string, l []string) bool {
	for _, s2 := range l {
		if s == s2 {
			return true
		}
	}
	return false
}
