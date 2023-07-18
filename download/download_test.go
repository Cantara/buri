package download

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/version/filter"
)

type MockPackageRepo struct {
}

func mockZip(f *os.File) {
	//Create a new zip writer
	zipWriter := zip.NewWriter(f)
	fmt.Println("opening first file")
	//Add files to the zip archive
	f1, err := os.Open("download.go")
	if err != nil {
		panic(err)
	}
	defer f1.Close()

	fmt.Println("adding file to archive..")
	w1, err := zipWriter.Create("download.go")
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(w1, f1); err != nil {
		panic(err)
	}
	fmt.Println("closing archive")
	zipWriter.Close()
}

func (pr MockPackageRepo) DownloadFile(dir, path, fileName string) (fullNewFilePath string) {
	log.Info("MOCK: Downloading new version", "name", fileName, "path", path)

	fullNewFilePath = filepath.Clean(fmt.Sprintf("%s/%s", dir, fileName))
	out, err := os.OpenFile(fullNewFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.WithError(err).Fatal("while opening file to write download to")
	}
	defer out.Close()
	if strings.HasSuffix(fileName, ".zip") {
		mockZip(out)
	}
	return
}

func (pr MockPackageRepo) NewestVersion(diskFS fs.FS, f filter.Filter, groupId, artifactId, linkName, packageType, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error) {
	runtime.Gosched()
	return "", "1.0.0", false, nil //This moch will probably fail
}

func TestClean(t *testing.T) {
	dfs := os.DirFS(".")
	fs.WalkDir(dfs, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			return nil
		}
		os.Remove(path)
		return nil
	})
}

func TestDownload(t *testing.T) {
	pr := MockPackageRepo{}
	dir := "."
	artifactId := "testArtifact"
	linkName := artifactId
	groupId := "testGroup"
	repoUrl := ""
	subArtifact := []string{artifactId}
	f := filter.AllReleases

	for _, packageType := range []string{"go", "jar", "tar", "zip"} {
		newFileName := ArtifactDownloader{}.
			Download(os.DirFS(dir), pr, packageType, linkName, artifactId, groupId, repoUrl, subArtifact, f)
		if newFileName == "" {
			t.Errorf("Package Type %s is not downloadable!", packageType)
			continue
		}
		if strings.Count(newFileName, packageType) > 1 {
			t.Errorf("New file name(%s) contains the PackageType(%s) more than once.", newFileName, packageType)
		}
		if packageType == "tar" || packageType == "zip" {
			if strings.Contains(newFileName, runtime.GOARCH) {
				t.Errorf("New file name for packaged filetyle contains arch in name")
				continue
			}
			if strings.Contains(newFileName, runtime.GOOS) {
				t.Errorf("New file name for packaged filetyle contains os in name")
				continue
			}
		}
	}
}
