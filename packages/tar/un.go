package tar

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/cantara/bragi"
)

func Unpack(srcFile string) (err error) {
	defer func() {
		if errors.Is(err, io.EOF) {
			err = nil
		}
	}()
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
