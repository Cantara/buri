package disk

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/readers"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic/parser"
)

func Version[T readers.Version[T]](disk fs.FS, f filter.Filter, linkName string, packageType pack.Type, numVersionsToKeep int) (versionsOnDisk []readers.Program[T], runningVersion T, removeLink bool, err error) {
	err = fs.WalkDir(disk, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		log.Debug("reading", "path", path)
		if path == linkName {
			linkPath, err := os.Readlink(fmt.Sprintf("%s/%s", disk, path))
			log.Trace("read link", "path", path, "link", linkPath)
			if err != nil {
				log.WithError(err).Error("while trying to read symlink to get current version")
				return nil
			}
			linkPathParts := strings.Split(linkPath, "/")
			linkPath = linkPathParts[len(linkPathParts)-1]
			artifactId := linkName
			fileName := strings.ReplaceAll(strings.ReplaceAll(filepath.Base(linkPath), "-"+runtime.GOOS, ""), "-"+runtime.GOARCH, "")
			//fileNameEls := strings.Split(fileName, "-")
			runningVersionString := fileName
			switch packageType {
			case pack.Jar:
				runningVersionString = strings.TrimSuffix(runningVersionString, ".jar")
				artifactId = strings.TrimSuffix(artifactId, ".jar")
			case pack.Tar:
				runningVersionString = strings.TrimSuffix(runningVersionString, ".tgz")
				artifactId = strings.TrimSuffix(artifactId, ".tgz")
			case pack.Zip:
				runningVersionString = strings.TrimSuffix(runningVersionString, ".zip")
				artifactId = strings.TrimSuffix(artifactId, ".zip")
			}
			runningVersionString = strings.TrimPrefix(runningVersionString, artifactId+"-")
			log.Trace("modified link", "path", path, "link", linkPath, "artifactId", artifactId, "filename", fileName, "version", runningVersionString)
			runningVersionAny, err := parser.Parse(f, runningVersionString)
			if err != nil {
				log.WithError(err).Debug("while trying to parse version")
				return nil
			}
			runningVersion = runningVersionAny.(T)
			removeLink = true
			return nil
		}
		if strings.HasPrefix(path, linkName+"-") {
			log.Debug(path)
			log.Debug(linkName)
			name := filepath.Base(path)
			name = strings.TrimSuffix(name, ".jar")
			name = strings.TrimSuffix(name, ".zip")
			name = strings.TrimSuffix(name, ".tgz")
			name = strings.ReplaceAll(name, "-"+runtime.GOOS, "")
			name = strings.ReplaceAll(name, "-"+runtime.GOARCH, "")
			name = strings.ReplaceAll(name, "-SNAPSHOT", "")
			name = strings.ReplaceAll(name, "-RC", "")
			versionString := strings.ReplaceAll(name, linkName+"-", "")
			if len(strings.Split(versionString, "-")) > 1 {
				log.Debug("skipping since it is probably a sub artifact: ", versionString)
				return nil
			}
			vers, err := parser.Parse(f, versionString)
			if err != nil {
				return nil
			}
			versionsOnDisk = append(versionsOnDisk, readers.Program[T]{
				Path:    path,
				Version: vers.(T),
			})
		}
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	})

	log.Trace("current disk status", "versions", versionsOnDisk)

	sort.Slice(versionsOnDisk, func(i, j int) bool {
		return versionsOnDisk[i].Version.IsStrictlySemanticNewer(f, versionsOnDisk[j].Version)
	})
	defer func() {
		for i := 0; i < len(versionsOnDisk)-numVersionsToKeep; i++ {
			f := filepath.Clean(fmt.Sprintf("%s/%s", disk, versionsOnDisk[i].Path))
			err = os.RemoveAll(f)
			log.WithError(err).Info("removing", "dir", disk, "file", versionsOnDisk[i].Path)
		}
	}()
	return
}
