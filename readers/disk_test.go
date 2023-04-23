package readers

import (
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"os"
	"testing"
)

var disk = os.DirFS("test")

func TestVersionOnDisk(t *testing.T) {
	f, _ := filter.Parse("*.*.*")
	versionsOnDisk, runningVersion, removeLink, err := VersionOnDisk[release.Version](disk, f, "buri", "go")
	if err != nil {
		t.Fatal(err)
	}
	if !removeLink {
		t.Fatal("remove link should be true")
	}
	if runningVersion.Major != 0 || runningVersion.Minor != 9 || runningVersion.Patch != 8 {
		t.Fatal("incorrect running version")
	}
	if len(versionsOnDisk) != 1 {
		t.Fatal("number of running versions is incorrect")
	}
}
