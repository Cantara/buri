package release

import (
	"errors"
	"fmt"
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

func (v1 Version) IsSemanticNewer(filter filter.Filter, v2 Version) (newer bool, err error) {
	if !v1.Matches(filter) {
		err = ErrVersionDoesNotMatchFilter
		return
	}
	if !v2.Matches(filter) {
		err = ErrVersionDoesNotMatchFilter
		return
	}
	if v1.Major < v2.Major {
		newer = true
		return
	}
	if v1.Minor < v2.Minor {
		newer = true
		return
	}
	if v1.Patch < v2.Patch {
		newer = true
		return
	}
	return
}

func (v1 Version) IsStrictlySemanticNewer(filter filter.Filter, v2 Version) bool {
	if !v2.Matches(filter) {
		return false
	}
	if !v1.Matches(filter) {
		return true
	}
	newer, err := v1.IsSemanticNewer(filter, v2)
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
