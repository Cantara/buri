package snapshot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type FilterLevel int

func (fl FilterLevel) Locked(l FilterLevel) bool {
	return fl <= l
}

const (
	Free     = FilterLevel(4)
	Major    = FilterLevel(3)
	Minor    = FilterLevel(2)
	Patch    = FilterLevel(1)
	Snapshot = FilterLevel(0)
)

type Filter struct {
	Level   FilterLevel
	Version Version
}

func PatternToFilter(pattern string) (filter Filter, err error) {
	parts := strings.Split(pattern, ".")
	if len(parts) > 3 || len(parts) == 0 {
		err = ErrNotValidPattern
		return
	}

	versionSet := false
	freeParts := 0
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "*" {
			freeParts++
			if versionSet {
				err = ErrNotValidPattern
				return
			}
			continue
		}
		versionSet = true
		var vers int
		vers, err = strconv.Atoi(parts[i])
		if err != nil {
			err = errors.Join(err, ErrNotValidPattern)
			return
		}
		switch i {
		case 2:
			filter.Version.Major = vers
		case 1:
			filter.Version.Minor = vers
		case 0:
			filter.Version.Patch = vers
		}
	}
	filter.Level = FilterLevel(freeParts + 1)
	return
}

func IsSemanticNewer(filter Filter, v1, v2 SnapshotVersion) (newer bool, err error) {
	if filter.Level.Locked(Major) {
		if filter.Version.Major != v1.Version.Major {
			err = ErrVersionDoesNotMatchFilter
			return
		}
		if filter.Version.Major != v2.Version.Major {
			err = ErrVersionDoesNotMatchFilter
			return
		}
	}
	if filter.Level.Locked(Minor) {
		if filter.Version.Minor != v1.Version.Minor {
			err = ErrVersionDoesNotMatchFilter
			return
		}
		if filter.Version.Minor != v2.Version.Minor {
			err = ErrVersionDoesNotMatchFilter
			return
		}
	}
	if filter.Level.Locked(Patch) {
		if filter.Version.Patch != v1.Version.Patch {
			err = ErrVersionDoesNotMatchFilter
			return
		}
		if filter.Version.Patch != v2.Version.Patch {
			err = ErrVersionDoesNotMatchFilter
			return
		}
	}
	if v1.Version.Major < v2.Version.Major {
		newer = true
		return
	}
	if v1.Version.Minor < v2.Version.Minor {
		newer = true
		return
	}
	if v1.Version.Patch < v2.Version.Patch {
		newer = true
		return
	}
	if v1.TimeStamp.Before(v2.TimeStamp) {
		newer = true
		return
	}
	if v1.Iteration < v2.Iteration {
		newer = true
		return
	}
	return
}

func ParseSnapshotVersion(s string) (sv SnapshotVersion, err error) {
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		err = fmt.Errorf("err: %v, %s", ErrNotValidVersion, "snapshot version string did not have the correct format")
		return
	}
	vers, err := ParseVersion(parts[0])
	if err != nil {
		err = fmt.Errorf("err: %v, %s", err, "while parsing version")
		return
	}
	t, err := time.Parse("20060102.150405", parts[1])
	if err != nil {
		err = fmt.Errorf("err: %v, %s", err, "while parsing timestamp")
		return
	}
	itr, err := strconv.Atoi(parts[2])
	if err != nil {
		err = fmt.Errorf("err: %v, %s", err, "while parsing version iteration")
		return
	}
	sv = SnapshotVersion{
		Version:   vers,
		TimeStamp: t,
		Iteration: itr,
	}
	return
}

func GenerateSnapshotVersion(cur Version, itr int) SnapshotVersion {
	itr++
	return SnapshotVersion{
		Version:   cur,
		TimeStamp: time.Now(),
		Iteration: itr,
	}
}

type SnapshotVersion struct {
	Version   Version
	TimeStamp time.Time
	Iteration int
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(s string) (v Version, err error) {
	s = strings.TrimPrefix(s, "v")
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
	}
	return
}

var ErrNotValidVersion = errors.New("not a valid version")
var ErrNotValidPattern = errors.New("not a valid pattern")
var ErrVersionDoesNotMatchFilter = errors.New("version does not match filter")
