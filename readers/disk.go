package readers

import (
	"fmt"
	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/version/filter"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func VersionOnDisk[T Version[T]](disk fs.FS, f filter.Filter, linkName, packageType string) (versionsOnDisk []Program[T], runningVersion T, removeLink bool, err error) {
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
			runningVersionAny, err := ParseVersion(f, runningVersionString)
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
			versionString := strings.ReplaceAll(name, linkName+"-", "")
			if len(strings.Split(versionString, "-")) > 1 {
				log.Debug("skipping since it is probably a sub artifact: ", versionString)
				return nil
			}
			vers, err := ParseVersion(f, versionString)
			if err != nil {
				return err
			}
			versionsOnDisk = append(versionsOnDisk, Program[T]{
				Path:    path,
				Version: vers.(T),
			})
		}
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	})
	return
}
