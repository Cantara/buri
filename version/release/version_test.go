package release

import (
	"github.com/cantara/buri/version/filter"
	"testing"
)

func TestParseVersion(t *testing.T) {
	vers := "1.1.1"
	version, err := Parse(vers)
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
	f, _ := filter.Parse(pattern)
	v1, _ := Parse("2.1.9")
	v2, _ := Parse("2.1.10")
	newer, err := v2.IsSemanticNewer(f, v1)
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

func TestIsSemanticNewerLockedMinor(t *testing.T) {
	pattern := "2.1.*"
	f, _ := filter.Parse(pattern)
	v1, _ := Parse("2.1.9")
	v2, _ := Parse("2.1.10")
	newer, err := v1.IsSemanticNewer(f, v1)
	if err != nil {
		t.Fatal(err, f, v1, v2)
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

func TestIsStrictlySemanticNewer(t *testing.T) {
	pattern := "0.5.*"
	f, _ := filter.Parse(pattern)
	v1, _ := Parse("v0.6.34")
	v2, _ := Parse("v0.5.10")

	newer := v1.IsStrictlySemanticNewer(f, v1)
	if newer {
		t.Fatal("version was newer", "filter", f, "v1", v1)
	}
	newer = v2.IsStrictlySemanticNewer(f, v2)
	if newer {
		t.Fatal("version was newer", "filter", f, "v2", v2)
	}
	newer = v2.IsStrictlySemanticNewer(f, v1)
	if newer {
		t.Fatal("version was newer", "filter", f, "v1", v1, "v2", v2)
	}
	newer = v1.IsStrictlySemanticNewer(f, v2)
	if !newer {
		t.Fatal("version was not newer", "filter", f, "v1", v1, "v2", v2)
	}

	v1, _ = Parse("v0.1.0")
	v2, _ = Parse("v0.5.11")
	newer = v1.IsStrictlySemanticNewer(f, v2)
	if !newer {
		t.Fatal("version was not newer", "filter", f, "v1", v1, "v2", v2)
	}

	pattern = "*.*.*"
	f, _ = filter.Parse(pattern)
	v1, _ = Parse("v0.9.8")
	v2, _ = Parse("v0.10.1")
	newer = v1.IsStrictlySemanticNewer(f, v2)
	if !newer {
		t.Fatal("version was not newer", "filter", f, "v1", v1, "v2", v2)
	}
	v1, _ = Parse("v0.10.1")
	v2, _ = Parse("v0.9.8")
	newer = v1.IsStrictlySemanticNewer(f, v2)
	if newer {
		t.Fatal("version was newer", "filter", f, "v1", v1, "v2", v2)
	}
}
