package release

import (
	"errors"
	"fmt"
	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/version"
	"github.com/cantara/buri/version/filter"
	"strconv"
	"strings"
)

type Style int

const (
	GoStyle = Style(1)
)

const Type = version.Type("release")

type Version struct {
	Major int
	Minor int
	Patch int
	Style Style
}

func (v Version) Matches(f filter.Filter) bool {
	if f.Level.Locked(filter.Major) {
		if f.Version.Major != v.Major {
			return false
		}
	}
	if f.Level.Locked(filter.Minor) {
		if f.Version.Minor != v.Minor {
			return false
		}
	}
	if f.Level.Locked(filter.Patch) {
		if f.Version.Patch != v.Patch {
			return false
		}
	}
	return true
}

func (v Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)

	switch v.Style {
	case GoStyle:
		return "v" + version
	default:
		return version
	}
}

func (v Version) IsSemanticNewer(filter filter.Filter, v2 Version) (newer bool, err error) {
	if !v.Matches(filter) {
		err = ErrVersionDoesNotMatchFilter
		return
	}
	if !v2.Matches(filter) {
		err = ErrVersionDoesNotMatchFilter
		return
	}
	log.Trace("release is semantic newer", "filter", filter, "v1", v, "v2", v2)
	if v.Major < v2.Major {
		newer = true
		return
	}
	if v.Major > v2.Major {
		newer = false
		return
	}
	if v.Minor < v2.Minor {
		newer = true
		return
	}
	if v.Minor > v2.Minor {
		newer = false
		return
	}
	if v.Patch < v2.Patch {
		newer = true
		return
	}
	return
}

func (v Version) IsStrictlySemanticNewer(filter filter.Filter, v2 Version) bool {
	log.Trace("release testing strictly semantic newer", "filter", filter, "v1", v, "v2", v2)
	if !v2.Matches(filter) {
		return false
	}
	if !v.Matches(filter) {
		return true
	}
	newer, err := v.IsSemanticNewer(filter, v2)
	return newer && err == nil
}

func Parse(s string) (v Version, err error) {
	var style Style
	if strings.HasPrefix(s, "v") {
		style = GoStyle
		s = strings.TrimPrefix(s, "v")
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		err = ErrNotValidVersion
		return
	}
	var major int
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	var minor int
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	var patch int
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return
	}
	v = Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Style: style,
	}
	return
}

var ErrNotValidVersion = errors.New("not a valid version")
var ErrVersionDoesNotMatchFilter = errors.New("version does not match filter")
