package readers

import (
	"strings"

	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"github.com/cantara/buri/version/snapshot"
)

type VersionType interface {
	release.Version | snapshot.Version
}
type Version[T any] interface {
	VersionType
	IsStrictlySemanticNewer(f filter.Filter, v2 T) bool
	Matches(f filter.Filter) bool
	String() string
}

type Program[T Version[T]] struct {
	Path    string
	Version T
	//UpdatedTime time.Time
}

func (p Program[T]) DownloadPath() string {
	return strings.ReplaceAll(p.Path, "service/rest/repository/browse/", "repository/") + "/"
}
