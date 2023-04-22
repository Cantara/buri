package version

import (
	"errors"
	"strconv"
	"strings"
)

type FilterLevel int

func (fl FilterLevel) Locked(l FilterLevel) bool {
	return fl <= l
}

const (
	Free  = FilterLevel(4)
	Major = FilterLevel(3)
	Minor = FilterLevel(2)
	Patch = FilterLevel(1)
)

type Filter struct {
	Level    FilterLevel
	Version  Version
	Snapshot bool
}

func (f Filter) Matches(v Version) bool {
	if f.Level.Locked(Major) {
		if f.Version.Major != v.Major {
			return false
		}
	}
	if f.Level.Locked(Minor) {
		if f.Version.Minor != v.Minor {
			return false
		}
	}
	if f.Level.Locked(Patch) {
		if f.Version.Patch != v.Patch {
			return false
		}
	}
	return true
}

func ParseFilter(pattern string) (filter Filter, err error) {
	base := strings.Split(pattern, "-")
	if len(base) == 2 && strings.ToLower(base[1]) == "snapshot" {
		filter.Snapshot = true
	}
	parts := strings.Split(base[0], ".")
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
		case 0:
			filter.Version.Major = vers
		case 1:
			filter.Version.Minor = vers
		case 2:
			filter.Version.Patch = vers
		}
	}
	filter.Level = FilterLevel(freeParts + 1)
	return
}
