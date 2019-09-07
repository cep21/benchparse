package benchparse

// KeyValueList is an ordered list of possibly repeating key/value pairs where a key may happen more than once
type KeyValueList struct {
	keys []KeyValue
}

// AsMap returns the key value pairs as a map using only the first key's value if a key is duplicated in the list.
// Runs in O(N).
func (c KeyValueList) AsMap() map[string]string {
	ret := make(map[string]string)
	for _, k := range c.keys {
		if _, exists := ret[k.Key]; !exists {
			ret[k.Key] = k.Value
		}
	}
	return ret
}

// LookupAll returns all values for a single key.  Runs in O(N).
func (c KeyValueList) LookupAll(key string) []string {
	ret := make([]string, 0, 1)
	for _, k := range c.keys {
		if k.Key == key {
			ret = append(ret, k.Value)
		}
	}
	return ret
}

// Lookup a single key's value.  Returns false if the key does not exist (to distinguish from valid keys without values).
// Runs in O(N)
func (c KeyValueList) Lookup(key string) (string, bool) {
	for _, k := range c.keys {
		if k.Key == key {
			return k.Value, true
		}
	}
	return "", false
}

// Get a single key's value.  Returns empty string if the key does not exist.  If you want to know if the key existed,
// use Lookup or LookupAll.
func (c KeyValueList) Get(key string) string {
	ret, _ := c.Lookup(key)
	return ret
}

// KeyValue is a pair of key + value
type KeyValue struct {
	// The key of Key value pair
	Key string
	// The Value of key value pair
	Value string
}

func (k KeyValue) String() string {
	if k.Value == "" {
		return k.Key + ":"
	}
	return k.Key + ": " + k.Value
}
