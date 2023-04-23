package disk

import (
	"fmt"
	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/readers"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic/parser"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func Version[T readers.Version[T]](disk fs.FS, f filter.Filter, linkName, packageType string, numVersionsToKeep int) (versionsOnDisk []readers.Program[T], runningVersion T, removeLink bool, err error) {
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
			fileName := strings.ReplaceAll(strings.ReplaceAll(filepath.Base(linkPath), "-"+runtime.GOOS, ""), "-"+runtime.GOARCH, "")
			fileNameEls := strings.Split(fileName, "-")
			runningVersionString := fileNameEls[len(fileNameEls)-1]
			if strings.HasSuffix(packageType, "jar") { //should probably just do this always
				runningVersionString = strings.TrimSuffix(runningVersionString, ".jar")
			}
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
			err = os.RemoveAll(versionsOnDisk[i].Path)
			log.WithError(err).Info("while removing ", versionsOnDisk[i].Path)
		}
	}()
	return
}
