package readers

import (
	"github.com/cantara/buri/version"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"github.com/cantara/buri/version/snapshot"
)

func ParseVersion(f filter.Filter, s string) (vers any, err error) {
	switch f.Type {
	case snapshot.Type:
		vers, err = snapshot.Parse(s)
	case release.Type:
		vers, err = release.Parse(s)
	default:
		err = version.ErrTypeDoesNotExist
		return
	}
	return
}
