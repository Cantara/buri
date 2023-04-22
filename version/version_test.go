package version

import (
	"testing"
)

func TestParseFilter(t *testing.T) {
	{
		pattern := "*.*.*"
		filter, err := ParseFilter(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Free {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
	}
	{
		pattern := "1.*.*"
		filter, err := ParseFilter(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Major {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
	}
	{
		pattern := "*.2.*"
		_, err := ParseFilter(pattern)
		if err == nil {
			t.Fatal("pattern should not be valid")
		}
	}
	{
		pattern := "1.2.*"
		filter, err := ParseFilter(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Minor {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
	}
	{
		pattern := "1.2.5"
		filter, err := ParseFilter(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if filter.Level != Patch {
			t.Fatal("filter level is not correct", "level", filter.Level)
		}
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
	filter, _ := ParseFilter(pattern)
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

func TestIsSemanticNewerLockedMinor(t *testing.T) {
	pattern := "2.1.*"
	filter, _ := ParseFilter(pattern)
	v1, _ := ParseVersion("2.1.9")
	v2, _ := ParseVersion("2.1.10")
	newer, err := IsSemanticNewer(filter, v1, v1)
	if err != nil {
		t.Fatal(err, filter, v1, v2)
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

func TestIsStrictlySemanticNewer(t *testing.T) {
	pattern := "0.5.*"
	filter, _ := ParseFilter(pattern)
	v1, _ := ParseVersion("v0.6.34")
	v2, _ := ParseVersion("v0.5.10")

	newer := IsStrictlySemanticNewer(filter, v1, v1)
	if newer {
		t.Fatal("version was newer", "filter", filter, "v1", v1)
	}
	newer = IsStrictlySemanticNewer(filter, v2, v2)
	if newer {
		t.Fatal("version was newer", "filter", filter, "v2", v2)
	}
	newer = IsStrictlySemanticNewer(filter, v2, v1)
	if newer {
		t.Fatal("version was newer", "filter", filter, "v1", v1, "v2", v2)
	}
	newer = IsStrictlySemanticNewer(filter, v1, v2)
	if !newer {
		t.Fatal("version was not newer", "filter", filter, "v1", v1, "v2", v2)
	}

	v1, _ = ParseVersion("v0.1.0")
	v2, _ = ParseVersion("v0.5.11")
	newer = IsStrictlySemanticNewer(filter, v1, v2)
	if !newer {
		t.Fatal("version was not newer", "filter", filter, "v1", v1, "v2", v2)
	}

}
