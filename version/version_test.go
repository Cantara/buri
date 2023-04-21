package version

import (
	"testing"
)

func TestPatternToFilter(t *testing.T) {
	pattern := "*.*.*"
	filter, err := PatternToFilter(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if filter.Level != Free {
		t.Fatal("filter level is not correct", "level", filter.Level)
	}
}

func TestParseVersion(t *testing.T) {
	vers := "1.1.1"
	version, err := ParseVersion(vers)
	if err != nil {
		t.Fatal(err)
	}
	if version.Major != 1 {
		t.Fatal("major version is wrong", "major", version.Major)
	}
	if version.Minor != 1 {
		t.Fatal("minor version is wrong", "minor", version.Minor)
	}
	if version.Patch != 1 {
		t.Fatal("patch version is wrong", "patch", version.Patch)
	}
}

func TestIsSemanticNewer(t *testing.T) {
	pattern := "*.*.*"
	filter, _ := PatternToFilter(pattern)
	v1, _ := ParseVersion("2.1.9")
	v2, _ := ParseVersion("2.1.10")
	newer, err := IsSemanticNewer(filter, v1, v1)
	if err != nil {
		t.Fatal(err)
	}
	if newer {
		t.Fatal("version was newer", "filter", filter, "v1", v1)
	}
	newer, err = IsSemanticNewer(filter, v2, v2)
	if err != nil {
		t.Fatal(err)
	}
	if newer {
		t.Fatal("version was newer", "filter", filter, "v2", v2)
	}
	newer, err = IsSemanticNewer(filter, v2, v1)
	if err != nil {
		t.Fatal(err)
	}
	if newer {
		t.Fatal("version was newer", "filter", filter, "v1", v1, "v2", v2)
	}
	newer, err = IsSemanticNewer(filter, v1, v2)
	if err != nil {
		t.Fatal(err)
	}
	if !newer {
		t.Fatal("version was not newer", "filter", filter, "v1", v1, "v2", v2)
	}
}
