package shortcodes

type Attributes struct {
	Keys   []string
	Lookup map[string]string
}

func newNodeAttributes() (attrs *Attributes) {
	attrs = &Attributes{
		Keys:   make([]string, 0),
		Lookup: make(map[string]string),
	}
	return
}

func (a *Attributes) Set(key string, value string) {
	a.Keys = append(a.Keys, key)
	a.Lookup[key] = value
}

func (a *Attributes) Append(key string, value string) {
	if _, present := a.Lookup[key]; present {
		a.Lookup[key] += value
	} else {
		a.Set(key, value)
	}
}

func (a *Attributes) String() (output string) {
	for _, key := range a.Keys {
		output += " " + key + "=" + a.Lookup[key]
	}
	return
}