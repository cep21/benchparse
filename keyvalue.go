package benchparse

// OrderedStringStringMap is a map of strings to strings that maintains ordering.
// This statement implies uniqueness of keys per benchmark.
// "The interpretation of a key/value pair is up to tooling, but the key/value pair is considered to describe all benchmark results that follow, until overwritten by a configuration line with the same key."
type OrderedStringStringMap struct {
	// Contents are the values inside this map
	Contents map[string]string
	// Order is the string order of the contents of this map.  It is intended that len(Order) == len(Contents) and the
	// keys of Contents are all inside Order.
	Order []string
}

// valuesToTransition returns the OrderedStringStringMap object that is required to transition from the current
// key/value pairs to the newState of key/value pairs.  Not all transitions are possible.  It does a best guess
// ordering.
func (o *OrderedStringStringMap) valuesToTransition(newState *OrderedStringStringMap) *OrderedStringStringMap {
	if o == newState {
		return &OrderedStringStringMap{}
	}
	if o == nil || len(o.Contents) == 0 {
		return newState
	}
	if newState == nil {
		return o
	}
	ret := &OrderedStringStringMap{}
	for _, k := range newState.Order {
		v := newState.Contents[k]
		if !o.exists(k, v) {
			ret.add(k, v)
		}
	}
	return ret
}

func (o *OrderedStringStringMap) clone() OrderedStringStringMap {
	ret := OrderedStringStringMap{}
	for i := range o.Order {
		ret.add(o.Order[i], o.Contents[o.Order[i]])
	}
	return ret
}

func (o *OrderedStringStringMap) exists(k string, v string) bool {
	return o.Contents[k] == v
}

func (o *OrderedStringStringMap) add(k string, v string) {
	if _, exists := o.Contents[k]; exists {
		o.remove(k)
	}
	if o.Contents == nil {
		o.Contents = make(map[string]string)
	}
	o.Contents[k] = v
	o.Order = append(o.Order, k)
}

func (o *OrderedStringStringMap) remove(s string) {
	if _, exists := o.Contents[s]; !exists {
		return
	}
	delete(o.Contents, s)
	for i, val := range o.Order {
		if s == val {
			o.Order = append(o.Order[0:i], o.Order[i+1:]...)
			return
		}
	}
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
