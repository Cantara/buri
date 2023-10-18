package unpack

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/packages/tar"
	"github.com/cantara/buri/packages/zip"

	log "github.com/cantara/bragi/sbragi"
)

type Unpacker struct {
}

func (_ Unpacker) Unpack(localFS fs.FS, filePath, linkName string, packageType pack.Type) {
	dir := fmt.Sprint(localFS)
	os.Remove(filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName)))

	// Create the file
	log.Trace("unpacking new version", "os", runtime.GOOS, "arch", runtime.GOARCH, "packageType", packageType, "file", filePath)
	if packageType == pack.Tar {
		err := tar.Unpack(filePath)
		if err != nil {
			log.WithError(err).Fatal("while unpacking tar")
		}
		os.Remove(filePath)
		filePath = strings.TrimSuffix(filePath, ".tgz")
		linkName = strings.TrimSuffix(linkName, ".tgz")
	} else if packageType == pack.Zip {
		err := zip.Unpack(filePath)
		if err != nil {
			log.WithError(err).Fatal("while unpacking zip")
		}
		os.Remove(filePath)
		filePath = strings.TrimSuffix(filePath, ".zip")
		linkName = strings.TrimSuffix(linkName, ".zip")
	}
	fullLink := filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName))
	os.Remove(fullLink)
	err := os.Symlink(filePath, fullLink)
	if err != nil {
		log.WithError(err).Fatal("while sym linking")
	}
	return
}
