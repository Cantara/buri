package snapshot

import (
	"errors"
	"fmt"
	"github.com/cantara/buri/version"
	"strconv"
	"strings"
	"time"
)

func IsSemanticNewer(filter version.Filter, v1, v2 Version) (newer bool, err error) {
	newer, err = version.IsSemanticNewer(filter, v1.Version, v2.Version)
	if err != nil || newer {
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

func ParseSnapshotVersion(s string) (sv Version, err error) {
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		err = fmt.Errorf("err: %v, %s", ErrNotValidVersion, "snapshot version string did not have the correct format")
		return
	}
	vers, err := version.ParseVersion(parts[0])
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
	sv = Version{
		Version:   vers,
		TimeStamp: t,
		Iteration: itr,
	}
	return
}

func GenerateSnapshotVersion(cur version.Version, itr int) Version {
	itr++
	return Version{
		Version:   cur,
		TimeStamp: time.Now(),
		Iteration: itr,
	}
}

type Version struct {
	Version   version.Version
	TimeStamp time.Time
	Iteration int
}

var ErrNotValidVersion = errors.New("not a valid version")
