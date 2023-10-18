package cmd

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
	"github.com/cantara/buri/download"
	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/version/filter"
)

type MockConfigHandler struct {
}

func (_ MockConfigHandler) Config(artifactName string) (repoUrl string, f filter.Filter) {
	return "", filter.AllReleases
}

type MockPackageRepo struct {
}

func mockZip(f *os.File) {
	//Create a new zip writer
	zipWriter := zip.NewWriter(f)
	fmt.Println("opening first file")
	//Add files to the zip archive
	f1, err := os.Open("root.go")
	if err != nil {
		panic(err)
	}
	defer f1.Close()

	fmt.Println("adding file to archive..")
	w1, err := zipWriter.Create("root.go")
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
	out, err := os.OpenFile(filepath.Clean(fmt.Sprintf("%s/%s", dir, fileName)), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.WithError(err).Fatal("while opening file to write download to")
	}
	defer out.Close()
	if strings.HasSuffix(fileName, ".zip") {
		mockZip(out)
	}
	return
}

func (pr MockPackageRepo) NewestVersion(diskFS fs.FS, f filter.Filter, groupId, artifactId, linkName string, packageType pack.Type, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error) {
	runtime.Gosched()
	return "", "1.0.0", true, nil //This moch will probably fail
}

func TestClean(t *testing.T) {
	//Setting up debugLogger for tests
	debugLogger, _ := log.NewDebugLogger()
	debugLogger.SetDefault()

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

var globalT *testing.T

type MockInstallDownloader struct {
}

func (_ MockInstallDownloader) Download(localFS fs.FS, pr download.PackageRepo, packageType pack.Type, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (newFileName string) {
	if fmt.Sprint(localFS) != "/usr/local/bin" {
		globalT.Errorf("install directory not set correctly for packageType(%s). dir=\"%s\"", packageType, localFS)
	}
	installDir := "installTest"
	os.Mkdir(installDir, 0750)
	return download.ArtifactDownloader{}.Download(os.DirFS(installDir), pr, packageType, linkName, artifactId, groupId, repoUrl, subArtifact, f)
}

type MockGetDownloader struct {
}

var wd, _ = os.Getwd()

func (_ MockGetDownloader) Download(localFS fs.FS, pr download.PackageRepo, packageType pack.Type, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (newFileName string) {
	if fmt.Sprint(localFS) != wd {
		globalT.Errorf("install directory not set correctly for packageType(%s). dir=\"%s\"", packageType, localFS)
	}
	installDir := "getTest"
	os.Mkdir(installDir, 0750)
	return download.ArtifactDownloader{}.Download(os.DirFS(installDir), pr, packageType, linkName, artifactId, groupId, repoUrl, subArtifact, f)
}

type MockGetUnpacker struct {
}

func (_ MockGetUnpacker) Unpack(_ fs.FS, _, _ string, _ pack.Type) {
}

var packageTypes = []pack.Type{pack.Go, pack.Jar, pack.Tar, pack.Zip}

func TestInstall(t *testing.T) {
	globalT = t
	md := MockInstallDownloader{}
	mpr := MockPackageRepo{}
	mch := MockConfigHandler{}

	artifactId := "testArtifact"
	groupId := "testGroup"

	for _, packageType := range packageTypes {
		install(md, mpr, mch, packageType, artifactId, groupId)
	}
}

func TestGet(t *testing.T) {
	globalT = t
	md := MockGetDownloader{}
	mpr := MockPackageRepo{}
	mch := MockConfigHandler{}
	mup := MockGetUnpacker{}

	artifactId := "testArtifact"
	groupId := "testGroup"

	for _, packageType := range packageTypes {
		get(md, mup, mpr, mch, packageType, artifactId, groupId, false)
	}
}
