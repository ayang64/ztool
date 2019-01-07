package nvlist

type List map[string]interface{}

func (l List) Find(target string) (interface{}, bool) {
	// search submaps for value.
	for k := range l {
		sl, t := l[k].(List)
		if t == false {
			continue
		}
		if v, found := sl.Find(target); found {
			return v, found
		}
	}

	// search current lvel for value
	for k := range l {
		if k != target {
			continue
		}
		return l[k], true
	}

	return nil, false
}
