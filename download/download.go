package download

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/readers/maven"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic"
	"github.com/cantara/buri/version/release"

	log "github.com/cantara/bragi/sbragi"
)

func Download(dir, packageType, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (foundNewerVersion bool) {
	command := []string{fmt.Sprintf("%s/%s", dir, linkName)}
	if strings.HasSuffix(packageType, "jar") {
		command = []string{"java", "-jar", command[0]}
	}
	diskFS := os.DirFS(dir)
	mavenPath, mavenVersion, removeLink, err := generic.NewestVersion(diskFS, f, groupId, artifactId, linkName, packageType, repoUrl, 4)
	if mavenVersion != "" {
		log.Info("new version found", "version", mavenVersion)
		foundNewerVersion = true
	}

	if !foundNewerVersion {
		log.Info("no new version found")
		return
	}
	// Create the file

	var path string
	if len(subArtifact) == 1 {
		path = mavenPath
	} else {
		path = fmt.Sprintf("%s/%s/", strings.TrimSuffix(mavenPath, "/"), strings.Join(subArtifact[1:], "/"))
	}
	path = fmt.Sprintf("%s%s-%s", path, artifactId, mavenVersion)
	fileName := fmt.Sprintf("%s-%s", strings.TrimSuffix(linkName, ".jar"), mavenVersion)
	switch strings.ToLower(packageType) {
	case "go":
		fileName = fmt.Sprintf("%s-%s-%s", fileName, runtime.GOOS, runtime.GOARCH)
		path = fmt.Sprintf("%s-%s-%s", path, runtime.GOOS, runtime.GOARCH)
	case "jar":
		fileName = fmt.Sprintf("%s.jar", fileName)
		path = fmt.Sprintf("%s.jar", path)
	case "tgz":
		fileName = fmt.Sprintf("%s.tgz", fileName)
		path = fmt.Sprintf("%s.tgz", path)
	case "zip_jar":
		fileName = fmt.Sprintf("%s.zip", fileName)
		path = fmt.Sprintf("%s.zip", path)

	}
	maven.DownloadFile(dir, path, fileName)
	if removeLink {
		err = os.Remove(filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName)))
		if err != nil {
			log.WithError(err).Fatal("while removing link")
		}
	}
	if packageType == "tgz" {
		pack.UnTGZ(fileName)
		os.Remove(fileName)
		fileName = strings.TrimSuffix(fileName, ".tgz")
		linkName = strings.TrimSuffix(linkName, ".tgz")
	} else if strings.HasPrefix(packageType, "zip") {
		linkName = strings.TrimSuffix(linkName, ".zip")
		fn, err := pack.UnZip(fileName)
		if err != nil {
			log.WithError(err).Fatal("while unpacking zip")
		}
		os.Remove(fileName)
		os.Remove(linkName)
		versionParts := strings.Split(mavenVersion, "-")
		innerVersion := versionParts[0]
		if f.Type != release.Type {
			innerVersion = fmt.Sprintf("%s-%s", innerVersion, strings.ToUpper(string(f.Type)))
		}
		err = os.Symlink(fmt.Sprintf("%s/%s-%s.jar", fn, artifactId, innerVersion), linkName)
		if err != nil {
			log.WithError(err).Fatal("while symlinking inner jar")
		}
		fileName = strings.TrimSuffix(fileName, ".zip")
		linkName = strings.TrimSuffix(linkName, ".jar")
	}
	fullLink := filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName))
	os.Remove(fullLink)
	err = os.Symlink(fileName, fullLink)
	if err != nil {
		log.WithError(err).Fatal("while sym linking")
	}
	return
}
