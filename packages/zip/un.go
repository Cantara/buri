package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/cantara/bragi"
)

func Unpack(srcFile string) (err error) {
	base := strings.TrimSuffix(srcFile, ".zip")
	os.Mkdir(base, 0750)
	r, err := zip.OpenReader(srcFile)
	if err != nil {
		return
	}
	defer r.Close()
	//fn = filepath.Base(srcFile)
	//fn = strings.TrimSuffix(fn, filepath.Ext(fn))
	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		osFileName := filepath.Clean(fmt.Sprintf("%s/%s", base, f.Name))
		if f.FileInfo().IsDir() {
			os.MkdirAll(osFileName, 0750)
			continue
		}
		func() {
			zf, err := f.Open()
			if err != nil {
				log.Fatal(err)
			}
			defer zf.Close()
			df, err := os.OpenFile(osFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
			if err != nil {
				log.Fatal(err)
			}
			written, err := io.Copy(df, zf)
			if err != nil {
				log.Fatal(err)
			}
			if written != f.FileInfo().Size() {
				log.Fatal("wrote less than file size")
			}
		}()
	}
	return
}
