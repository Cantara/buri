package snapshot

import (
	"errors"
	"fmt"
	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/version"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"strconv"
	"strings"
	"time"
)

const Type = version.Type("snapshot")
const TimeStampFormat = "20060102.150405"

func Parse(s string) (sv Version, err error) {
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		err = fmt.Errorf("err: %v, %s", ErrNotValidVersion, "snapshot version string did not have the correct format")
		return
	}
	vers, err := release.Parse(parts[0])
	if err != nil {
		err = fmt.Errorf("err: %v, %s", err, "while parsing version")
		return
	}
	t, err := time.Parse(TimeStampFormat, parts[1])
	if err != nil {
		err = fmt.Errorf("err: %v, %s", err, "while parsing timestamp")
		return
	}
	itr, err := strconv.Atoi(parts[2])
	if err != nil {
		err = fmt.Errorf("err: %v, %s", err, "while parsing version iteration")
		return
	}
	sv = Version{
		Version:   vers,
		TimeStamp: t,
		Iteration: itr,
	}
	return
}

func Generate(cur release.Version, itr int) Version {
	itr++
	return Version{
		Version:   cur,
		TimeStamp: time.Now(),
		Iteration: itr,
	}
}

type Version struct {
	Version   release.Version
	TimeStamp time.Time
	Iteration int
}

func (v Version) Matches(f filter.Filter) bool {
	if !v.Version.Matches(f) {
		return false
	}
	return true
}

func (v Version) IsSemanticNewer(filter filter.Filter, v2 Version) (newer bool, err error) {
	newer, err = v.Version.IsSemanticNewer(filter, v2.Version)
	if err != nil || newer {
		return
	}
	if v.TimeStamp.Before(v2.TimeStamp) {
		newer = true
		return
	}
	if v.Iteration < v2.Iteration {
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

func (v Version) String() string {
	return fmt.Sprintf("%s-%s-%d", v.Version, v.TimeStamp.Format(TimeStampFormat), v.Iteration)
}

var ErrNotValidVersion = errors.New("not a valid version")
