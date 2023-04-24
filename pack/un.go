package pack

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	log "github.com/cantara/bragi"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnTGZ(srcFile string) (err error) {
	base := strings.TrimSuffix(srcFile, ".tgz")
	os.Mkdir(base, 0750)
	tgz, err := os.Open(srcFile)
	if err != nil {
		return
	}
	defer tgz.Close()

	gzf, err := gzip.NewReader(tgz)
	if err != nil {
		return
	}

	tarReader := tar.NewReader(gzf)
	for {
		header, err := tarReader.Next()
		if err != nil {
			return err
		}

		name := header.Name
		fmt.Println(name)

		switch header.Typeflag {
		case tar.TypeDir:
			fmt.Println("Directory:", name)
			os.Mkdir(fmt.Sprintf("%s/%s", base, name), 0750)
		case tar.TypeReg:
			fmt.Println("Regular file:", name)
			func() {
				fn := fmt.Sprintf("%s/%s", base, name)
				f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
				if err != nil {
					log.AddError(err).Error("while opening file, ", name)
					return
				}
				defer f.Close()
				_, err = io.Copy(f, tarReader)
				if err != nil {
					log.AddError(err).Error("while reading file ", name)
					return
				}
			}()
		default:
			log.Warning("not a known file type, ", name, ", ", header.Typeflag)
		}
	}
}

func UnZip(srcFile string) (fn string, err error) {
	base := strings.TrimSuffix(srcFile, ".tgz")
	os.Mkdir(base, 0750)
	r, err := zip.OpenReader(srcFile)
	if err != nil {
		return
	}
	defer r.Close()
	fn = filepath.Base(srcFile)
	fn = strings.TrimSuffix(fn, filepath.Ext(fn))
	for _, f := range r.File {
		fmt.Printf("Contents of %s:\n", f.Name)
		osFileName := filepath.Clean(fmt.Sprintf("%s/%s", fn, f.Name))
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
