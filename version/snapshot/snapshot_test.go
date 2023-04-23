package snapshot

import (
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"testing"
)

func TestParseSnapshotVersion(t *testing.T) {
	vers := "0.16.5-20230418.055134-1"
	sv, err := Parse(vers)
	if err != nil {
		t.Fatal(err)
	}
	if sv.Version.Major != 0 {
		t.Fatal("major version is wrong", "major", sv.Version.Major)
	}
	if sv.Version.Minor != 16 {
		t.Fatal("minor version is wrong", "minor", sv.Version.Minor)
	}
	if sv.Version.Patch != 5 {
		t.Fatal("patch version is wrong", "patch", sv.Version.Patch)
	}
	if sv.TimeStamp.Year() != 2023 {
		t.Fatal("year was wrong", "year", sv.TimeStamp.Year())
	}
	if sv.TimeStamp.Month() != 4 {
		t.Fatal("month was wrong", "month", sv.TimeStamp.Month())
	}
	if sv.TimeStamp.Day() != 18 {
		t.Fatal("day was wrong", "day", sv.TimeStamp.Day())
	}
	if sv.TimeStamp.Hour() != 5 {
		t.Fatal("hour was wrong", "hour", sv.TimeStamp.Hour())
	}
	if sv.TimeStamp.Minute() != 51 {
		t.Fatal("minute was wrong", "minute", sv.TimeStamp.Minute())
	}
	if sv.TimeStamp.Second() != 34 {
		t.Fatal("second was wrong", "second", sv.TimeStamp.Second())
	}
	if sv.Iteration != 1 {
		t.Fatal("iteration was wrong", "iteration", sv.Iteration)
	}
}

func TestIsSemanticNewer(t *testing.T) {
	pattern := "*.*.*"
	f, _ := filter.Parse(pattern)
	v1, _ := Parse("2.1.9-20230409.123528-1")
	v2, _ := Parse("2.1.9-20230409.141416-2")
	newer, err := v1.IsSemanticNewer(f, v1)
	if err != nil {
		t.Fatal(err)
	}
	if newer {
		t.Fatal("version was newer", "filter", f, "v1", v1)
	}
	newer, err = v2.IsSemanticNewer(f, v2)
	if err != nil {
		t.Fatal(err)
	}
	if newer {
		t.Fatal("version was newer", "filter", f, "v2", v2)
	}
	newer, err = v2.IsSemanticNewer(f, v1)
	if err != nil {
		t.Fatal(err)
	}
	if newer {
		t.Fatal("version was newer", "filter", f, "v1", v1, "v2", v2)
	}
	newer, err = v1.IsSemanticNewer(f, v2)
	if err != nil {
		t.Fatal(err)
	}
	if !newer {
		t.Fatal("version was not newer", "filter", f, "v1", v1, "v2", v2)
	}
}

func TestGenerateSnapshotVersion(t *testing.T) {
	vers := release.Version{
		Major: 1,
		Minor: 5,
		Patch: 2,
	}
	sv := Generate(vers, 2)
	if sv.Version.Major != 1 {
		t.Fatal("wrong major version", "major", sv.Version.Major)
	}
	if sv.Version.Minor != 5 {
		t.Fatal("wrong minor version", "minor", sv.Version.Minor)
	}
	if sv.Version.Patch != 2 {
		t.Fatal("wrong patch version", "patch", sv.Version.Patch)
	}
	if sv.Iteration != 3 {
		t.Fatal("wrong iteration", "iteration", sv.Iteration)
	}
}
