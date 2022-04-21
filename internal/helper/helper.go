package helper

func StringInSlice(s string, l []string) bool {
	for _, s2 := range l {
		if s == s2 {
			return true
		}
	}
	return false
}

func Unique(in []string) (out []string) {
	keys := make(map[string]bool)
	for _, s := range in {
		if _, ok := keys[s]; !ok {
			keys[s] = true
			out = append(out, s)
		}
	}
	return
}
