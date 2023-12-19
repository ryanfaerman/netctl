package target

import "fmt"

type List map[string]Target

func NewList() List {
	return List{}
}

func (t List) All() []Target {
	list := []Target{}
	for _, target := range t {
		list = append(list, target)
	}
	return list
}

// Add a target to the list by specifying the GOOS and GOARCH.
func (t List) Add(goos, goarch string) {
	t[fmt.Sprintf("%s-%s", goos, goarch)] = New(goos, goarch)
}

// Add a target to the list.
func (t List) AddTarget(target Target) {
	t.Add(target.GOOS, target.GOARCH)
}

type TargetFilter func(Target) bool

func ByGoos(goos string) TargetFilter {
	return func(t Target) bool {
		return t.GOOS == goos
	}
}

func ByGoarch(goarch string) TargetFilter {
	return func(t Target) bool {
		return t.GOARCH == goarch
	}
}

// Filter returns a new list of targets that match the given filter function
func (t List) Filter(fn TargetFilter) List {
	l := NewList()
	for _, target := range t {
		if fn(target) {
			l.Add(target.GOOS, target.GOARCH)
		}
	}
	return l
}

// Each iterates over the list of targets and calls the given function for each
func (t List) Each(fn func(Target), filters ...TargetFilter) {
	list := t
	for _, f := range filters {
		list = list.Filter(f)
	}

	for _, target := range list {
		fn(target)
	}

}

// Remove all targets matching the given filter functions.
func (t List) RemoveFilter(fn ...TargetFilter) {
	for k, v := range t {
		matched := 0
		for _, f := range fn {
			if f(v) {
				matched += 1
			}
		}

		// all filters matched, remove the target
		if matched == len(fn) {
			delete(t, k)
		}
	}
}

// SUGAR TIME!

var defaultList = NewList()

func Add(goos, goarch string) {
	defaultList.Add(goos, goarch)
}
func AddTarget(target Target) {
	defaultList.AddTarget(target)
}

func All() []Target {
	return defaultList.All()
}

func Filter(fn TargetFilter) List {
	return defaultList.Filter(fn)
}

func Each(fn func(Target), filters ...TargetFilter) {
	defaultList.Each(fn, filters...)
}

func RemoveFilter(fn ...TargetFilter) {
	defaultList.RemoveFilter(fn...)
}
