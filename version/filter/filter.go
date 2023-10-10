package filter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/version"
)

var (
	AllReleases  = must(Parse("*.*.*"))
	AllSnapshots = must(Parse("*.*.*-SNAPSHOT"))
)

type Level int

func (fl Level) Locked(l Level) bool {
	return fl <= l
}

const (
	Free  = Level(4)
	Major = Level(3)
	Minor = Level(2)
	Patch = Level(1)
)

type Version struct {
	Major int
	Minor int
	Patch int
}

type Filter struct {
	Level   Level
	Version Version
	Type    version.Type
}

func (f Filter) Matches(s string) bool {
	f2, err := Parse(s)
	if err != nil {
		log.WithError(err).Trace("string matches filter", "string", s)
		return false
	}
	if f.Type != f2.Type {
		return false
	}

	return true
}

func (f Filter) String() string {
	base := ""
	switch f.Level {
	case Free:
		base = "*.*.*"
	case Major:
		base = fmt.Sprintf("%d.*.*", f.Version.Major)
	case Minor:
		base = fmt.Sprintf("%d.%d.*", f.Version.Major, f.Version.Minor)
	case Patch:
		base = fmt.Sprintf("%d.%d.%d", f.Version.Major, f.Version.Minor, f.Version.Patch)
	}
	if f.Type == version.Base {
		return base
	}
	return fmt.Sprintf("%s-%s", base, strings.ToUpper(string(f.Type)))
}

func Parse(pattern string) (filter Filter, err error) {
	pattern = strings.TrimPrefix(pattern, "v")
	base := strings.Split(pattern, "-")
	filter.Type = version.Base //This is a bit weird
	if len(base) == 2 {
		filter.Type = version.Type(strings.ToLower(base[1]))
	}
	log.Debug("filter parsing", "type", filter.Type, "base", base, "pattern", pattern)
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
	filter.Level = Level(freeParts + 1)
	return
}

func must(f Filter, err error) Filter {
	if err != nil {
		log.WithError(err).Fatal("while checking must condition for filter")
	}
	return f
}

var ErrNotValidPattern = errors.New("not a valid pattern")
