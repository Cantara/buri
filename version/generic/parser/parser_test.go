package parser

import (
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"github.com/cantara/buri/version/snapshot"
	"testing"
)

func TestParseVersionRelease(t *testing.T) {
	f, _ := filter.Parse("*.*.*")
	vers, err := Parse(f, "1.0.1")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := vers.(release.Version)
	if !ok {
		t.Fatal("version parsed was not release version")
	}
	if v.Major != 1 || v.Minor != 0 || v.Patch != 1 {
		t.Fatal("release version was incorrect")
	}
}

func TestParseVersionSnapshot(t *testing.T) {
	f, _ := filter.Parse("*.*.*-SNAPSHOT")
	vers, err := Parse(f, "2.1.9-20230409.123528-1")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := vers.(snapshot.Version)
	if !ok {
		t.Fatal("version parsed was not snapshot version")
	}
	if v.Version.Major != 2 || v.Version.Minor != 1 || v.Version.Patch != 9 {
		t.Fatal("release version was incorrect")
	}
}
