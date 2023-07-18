package maven

import (
	"testing"

	"github.com/cantara/buri/version/filter"
)

func TestNewestVersion(t *testing.T) {
	expected := "1.1.0"
	versions := []string{
		"1.0.0",
		"0.2.0",
		expected,
	}
	got := newestVersion(filter.AllReleases, "", versions)
	if got.String() != expected {
		t.Errorf("expected %s but got %s", expected, got)
	}
}
