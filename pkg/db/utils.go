package db

// compare two unordered slices of equal length and determine
// whether they contain the same elements

// assumming go's implementation of maps has ~O(1) access time,
// then this implementation runs in ~O(n)
func AreEqualSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// make a temp map to hash original slice (a) into keys to check for
	tmp := make(map[string]int, len(a))
	for _, s := range a {
		tmp[s] = 1
	}

	// iterate over b to check for existence of any of its elements
	for _, e := range b {
		if _, ok := tmp[e]; !ok {
			return false
		}
	}
	return true
}
