package target_test

import (
	"testing"

	"github.com/ryanfaerman/netctl/magefiles/target"
)

// TestListFilter tests the ListFilter function.
func TestListFilter(t *testing.T) {
	l := target.NewList()
	l.Add("linux", "amd64")
	l.Add("linux", "arm")
	l.Add("windows", "amd64")
	l.Add("darwin", "amd64")

	linux := l.Filter(target.ByGoos("linux"))
	if len(linux) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(linux))
	}

	amd64 := l.Filter(target.ByGoarch("amd64"))
	if len(amd64) != 3 {
		t.Errorf("Expected 3 targets, got %d", len(amd64))
	}
}

// TestListAdd tests the List.Add function.
func TestListAdd(t *testing.T) {
	l := target.NewList()
	l.Add("linux", "amd64")
	l.Add("linux", "arm")
	l.Add("windows", "amd64")
	l.Add("darwin", "amd64")

	if len(l) != 4 {
		t.Errorf("Expected 4 targets, got %d", len(l))
	}
}

// TestListAddTarget tests the List.AddTarget function.
func TestListAddTarget(t *testing.T) {
	l := target.NewList()
	l.AddTarget(target.New("linux", "amd64"))
	l.AddTarget(target.New("linux", "arm"))
	l.AddTarget(target.New("windows", "amd64"))
	l.AddTarget(target.New("darwin", "amd64"))

	if len(l) != 4 {
		t.Errorf("Expected 4 targets, got %d", len(l))
	}
}

// TestListRemoveFilter tests the List.RemoveFilter function.
func TestListRemoveFilter(t *testing.T) {
	l := target.NewList()
	l.Add("linux", "amd64")
	l.Add("linux", "arm")
	l.Add("windows", "amd64")
	l.Add("darwin", "amd64")

	l.RemoveFilter(target.ByGoos("linux"))

	if len(l) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(l))
	}
}

// TestListEach tests the List.Each function.
func TestListEach(t *testing.T) {
	l := target.NewList()
	l.Add("linux", "amd64")
	l.Add("linux", "arm")
	l.Add("windows", "amd64")
	l.Add("darwin", "amd64")

	count := 0
	l.Each(func(t target.Target) {
		count++
	})

	if count != 4 {
		t.Errorf("Expected 4 targets, got %d", count)
	}
}

// TestListEachFilter tests the List.Each function with filters.
func TestListEachFilter(t *testing.T) {
	l := target.NewList()
	l.Add("linux", "amd64")
	l.Add("linux", "arm")
	l.Add("windows", "amd64")
	l.Add("darwin", "amd64")

	count := 0
	l.Each(func(t target.Target) {
		count++
	}, target.ByGoos("linux"))

	if count != 2 {
		t.Errorf("Expected 2 targets, got %d", count)
	}
}
