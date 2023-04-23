package filter

import (
	"testing"
)

func TestParseFilter(t *testing.T) {
	{
		pattern := "*.*.*"
		filter, err := Parse(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Free {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
	}
	{
		pattern := "1.*.*"
		filter, err := Parse(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Major {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
	}
	{
		pattern := "*.2.*"
		_, err := Parse(pattern)
		if err == nil {
			t.Fatal("pattern should not be valid")
		}
	}
	{
		pattern := "1.2.*"
		filter, err := Parse(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Minor {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
	}
	{
		pattern := "1.2.5"
		filter, err := Parse(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Patch {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
	}
}
