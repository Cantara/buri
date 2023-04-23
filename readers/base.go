package readers

import (
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"github.com/cantara/buri/version/snapshot"
	"time"
)

type VersionType interface {
	release.Version | snapshot.Version
}
type Version[T any] interface {
	VersionType
	IsStrictlySemanticNewer(f filter.Filter, v2 T) bool
}

type Program[T Version[T]] struct {
	Path        string
	Version     T
	UpdatedTime time.Time
}
