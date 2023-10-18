package download

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/version/filter"

	log "github.com/cantara/bragi/sbragi"
)

type PackageRepo interface {
	DownloadFile(dir, path, filename string) (fullNewFilePath string)
	NewestVersion(localFS fs.FS, f filter.Filter, groupId, artifactId, linkName string, packageType pack.Type, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error)
}

type ArtifactDownloader struct {
}

func (_ ArtifactDownloader) Download(localFS fs.FS, pr PackageRepo, packageType pack.Type, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (fullNewFilePath string) {
	dir := fmt.Sprint(localFS)
	command := []string{fmt.Sprintf("%s/%s", localFS, linkName)}
	if packageType == pack.Jar {
		command = []string{"java", "-jar", command[0]}
	}
	var path string
	mavenPath, mavenVersion, removeLink, err := pr.NewestVersion(localFS, f, groupId, artifactId, linkName, packageType, repoUrl, 4)
	var newFileName string
	if mavenVersion != "" {
		log.Info("new version found", "version", mavenVersion)
		if len(subArtifact) == 1 {
			path = mavenPath
		} else {
			path = fmt.Sprintf("%s/%s/", strings.TrimSuffix(mavenPath, "/"), strings.Join(subArtifact[1:], "/"))
		}
		path = fmt.Sprintf("%s%s-%s", path, artifactId, mavenVersion)
		newFileName = fmt.Sprintf("%s-%s", packageType.TrimExtention(linkName), mavenVersion)
		switch packageType {
		case pack.Go:
			newFileName = fmt.Sprintf("%s-%s-%s", newFileName, runtime.GOOS, runtime.GOARCH)
			path = fmt.Sprintf("%s-%s-%s", path, runtime.GOOS, runtime.GOARCH)
		case pack.Jar:
			newFileName = fmt.Sprintf("%s.jar", newFileName)
			path = fmt.Sprintf("%s.jar", path)
		case pack.Tar:
			newFileName = fmt.Sprintf("%s.tgz", newFileName)
			path = fmt.Sprintf("%s.tgz", path)
		case pack.Zip:
			newFileName = fmt.Sprintf("%s.zip", newFileName)
			path = fmt.Sprintf("%s.zip", path)
		}
	}

	if newFileName == "" {
		log.Info("no new version found")
		return
	}
	// Create the file
	log.Trace("new version", "os", runtime.GOOS, "arch", runtime.GOARCH, "packageType", packageType, "file", newFileName)

	fullNewFilePath = pr.DownloadFile(dir, path, newFileName)
	fullLink := filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName))
	if removeLink {
		if err != nil {
			log.WithError(err).Warning("while removing link")
		}
		os.Remove(fullLink)
	}
	err = os.Symlink(fullNewFilePath, fullLink)
	if err != nil {
		log.WithError(err).Fatal("while sym linking")
	}
	return
}
