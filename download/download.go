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
	"github.com/cantara/buri/version/release"

	log "github.com/cantara/bragi/sbragi"
)

type PackageRepo interface {
	DownloadFile(dir, path, filename string)
	NewestVersion(localFS fs.FS, f filter.Filter, groupId, artifactId, linkName, packageType, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error)
}

func Download(localFS fs.FS, pr PackageRepo, packageType, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (newFileName string) {
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
		case "tgz":
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

	pr.DownloadFile(dir, path, newFileName)
	if removeLink {
		err = os.Remove(filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName)))
		if err != nil {
			log.WithError(err).Warning("while removing link")
		}
	}
	if packageType == "tgz" {
		pack.UnTGZ(newFileName)
		os.Remove(newFileName)
		newFileName = strings.TrimSuffix(newFileName, ".tgz")
		linkName = strings.TrimSuffix(linkName, ".tgz")
	} else if strings.HasPrefix(packageType, "zip") {
		linkName = strings.TrimSuffix(linkName, ".zip")
		err := pack.UnZip(newFileName)
		if err != nil {
			log.WithError(err).Fatal("while unpacking zip")
		}
		os.Remove(newFileName)
		os.Remove(linkName)
		versionParts := strings.Split(mavenVersion, "-")
		innerVersion := versionParts[0]
		if f.Type != release.Type {
			innerVersion = fmt.Sprintf("%s-%s", innerVersion, strings.ToUpper(string(f.Type)))
		}
		/*
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
	err = os.Symlink(newFileName, fullLink)
	if err != nil {
		log.WithError(err).Fatal("while sym linking")
	}
	return
}
