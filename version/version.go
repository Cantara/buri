package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func IsSemanticNewer(filter Filter, v1, v2 Version) (newer bool, err error) {
	if !filter.Matches(v1) {
		err = ErrVersionDoesNotMatchFilter
		return
	}
	if !filter.Matches(v2) {
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

func IsStrictlySemanticNewer(filter Filter, v1, v2 Version) bool {
	if !filter.Matches(v2) {
		return false
	}
	if !filter.Matches(v1) {
		return true
	}
	newer, err := IsSemanticNewer(filter, v1, v2)
	return newer && err == nil
}

type Style int

const (
	GoStyle = Style(1)
)

type Version struct {
	Major int
	Minor int
	Patch int
	Style Style
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

func (v Version) IsSemanticNewer(filter Filter, v2 Version) (newer bool, err error) {
	return IsSemanticNewer(filter, v, v2)
}

func (v Version) IsStrictlySemanticNewer(filter Filter, v2 Version) bool {
	newer, err := v.IsSemanticNewer(filter, v2)
	return newer && err == nil
}

func ParseVersion(s string) (v Version, err error) {
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
var ErrNotValidPattern = errors.New("not a valid pattern")
var ErrVersionDoesNotMatchFilter = errors.New("version does not match filter")
