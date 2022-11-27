package mirage

type StringSet map[string]struct{}

// NewStringSet returns a new string set
func NewStringSet(values []string) StringSet {
	ss := StringSet{}
	for _, v := range values {
		ss[v] = struct{}{}
	}

	return ss
}

// Has returns true if the value is present in the set
func (ss StringSet) Has(value string) bool {
	_, ok := ss[value]
	return ok
}
