package generic

import (
	"fmt"
	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/readers"
	"github.com/cantara/buri/readers/disk"
	"github.com/cantara/buri/readers/maven"
	"github.com/cantara/buri/version"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/release"
	"github.com/cantara/buri/version/snapshot"
	"io/fs"
)

func NewestVersion(diskFS fs.FS, f filter.Filter, groupId, artifactId, linkName, packageType, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error) {
	log.Info("finding newest version", "filter", f)
	switch f.Type {
	case snapshot.Type:
		mavenPath, mavenVersion, removeLink, err = newestVersion[snapshot.Version](diskFS, f, groupId, artifactId, linkName, packageType, repoUrl, numVersionsToKeep)
	case release.Type:
		mavenPath, mavenVersion, removeLink, err = newestVersion[release.Version](diskFS, f, groupId, artifactId, linkName, packageType, repoUrl, numVersionsToKeep)
	default:
		err = version.ErrTypeDoesNotExist
		return
	}
	return
}

func newestVersion[T readers.Version[T]](diskFS fs.FS, f filter.Filter, groupId, artifactId, linkName, packageType, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error) {
	var runningVersion T
	_, runningVersion, removeLink, err = disk.Version[T](diskFS, f, linkName, packageType, numVersionsToKeep)
	if err != nil {
		log.WithError(err).Fatal("finding disk version")
		return
	}

	url := fmt.Sprintf("%s/%s/%s", repoUrl, groupId, artifactId)
	versionInMavem := maven.Version[T](f, url, packageType)
	foundNewerVersion := runningVersion.IsStrictlySemanticNewer(f, versionInMavem.Version)
	log.Info("version check", "local", runningVersion, "maven", versionInMavem.Version, "isNew", foundNewerVersion)
	if !foundNewerVersion {
		return
	}
	mavenPath = versionInMavem.Path
	mavenVersion = versionInMavem.Version.String()
	return
}
