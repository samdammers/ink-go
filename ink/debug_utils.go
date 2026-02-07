package ink

func keys(m map[string]RuntimeObject) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}
