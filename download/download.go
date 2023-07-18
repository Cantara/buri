package download

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cantara/buri/packages/tar"
	"github.com/cantara/buri/packages/zip"
	"github.com/cantara/buri/version/filter"

	log "github.com/cantara/bragi/sbragi"
)

type PackageRepo interface {
	DownloadFile(dir, path, filename string) (fullNewFilePath string)
	NewestVersion(localFS fs.FS, f filter.Filter, groupId, artifactId, linkName, packageType, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error)
}

type ArtifactDownloader struct {
}

func (_ ArtifactDownloader) Download(localFS fs.FS, pr PackageRepo, packageType, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (newFileName string) {
	dir := fmt.Sprint(localFS)
	command := []string{fmt.Sprintf("%s/%s", localFS, linkName)}
	if strings.HasSuffix(packageType, "jar") {
		command = []string{"java", "-jar", command[0]}
	}
	var path string
	mavenPath, mavenVersion, removeLink, err := pr.NewestVersion(localFS, f, groupId, artifactId, linkName, packageType, repoUrl, 4)
	if mavenVersion != "" {
		log.Info("new version found", "version", mavenVersion)
		if len(subArtifact) == 1 {
			path = mavenPath
		} else {
			path = fmt.Sprintf("%s/%s/", strings.TrimSuffix(mavenPath, "/"), strings.Join(subArtifact[1:], "/"))
		}
		path = fmt.Sprintf("%s%s-%s", path, artifactId, mavenVersion)
		newFileName = fmt.Sprintf("%s-%s", strings.TrimSuffix(linkName, ".jar"), mavenVersion)
		switch strings.ToLower(packageType) {
		case "go":
			newFileName = fmt.Sprintf("%s-%s-%s", newFileName, runtime.GOOS, runtime.GOARCH)
			path = fmt.Sprintf("%s-%s-%s", path, runtime.GOOS, runtime.GOARCH)
		case "jar":
			newFileName = fmt.Sprintf("%s.jar", newFileName)
			path = fmt.Sprintf("%s.jar", path)
		case "tar":
			newFileName = fmt.Sprintf("%s.tgz", newFileName)
			path = fmt.Sprintf("%s.tgz", path)
		case "zip":
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

	fullNewFilePath := pr.DownloadFile(dir, path, newFileName)
	if removeLink {
		err = os.Remove(filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName)))
		if err != nil {
			log.WithError(err).Warning("while removing link")
		}
	}
	if packageType == "tar" {
		tar.Unpack(fullNewFilePath)
		os.Remove(fullNewFilePath)
		fullNewFilePath = strings.TrimSuffix(fullNewFilePath, ".tgz")
		linkName = strings.TrimSuffix(linkName, ".tgz")
	} else if packageType == "zip" {
		err := zip.Unpack(fullNewFilePath)
		if err != nil {
			log.WithError(err).Fatal("while unpacking zip")
		}
		os.Remove(fullNewFilePath)
		fullNewFilePath = strings.TrimSuffix(fullNewFilePath, ".zip")
		linkName = strings.TrimSuffix(linkName, ".zip")
		/*
			versionParts := strings.Split(mavenVersion, "-")
			innerVersion := versionParts[0]
			if f.Type != release.Type {
				innerVersion = fmt.Sprintf("%s-%s", innerVersion, strings.ToUpper(string(f.Type)))
			}
				err = os.Symlink(fmt.Sprintf("%s/%s-%s.jar", linkName, artifactId, innerVersion), linkName)
				if err != nil {
					log.WithError(err).Fatal("while symlinking inner jar")
				}
		*/
		//newFileName = strings.TrimSuffix(newFileName, ".zip")
		//linkName = strings.TrimSuffix(linkName, ".jar")
	}
	fullLink := filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName))
	os.Remove(fullLink)
	err = os.Symlink(fullNewFilePath, fullLink)
	if err != nil {
		log.WithError(err).Fatal("while sym linking")
	}
	return
}
