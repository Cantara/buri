package maven

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/cantara/bragi/sbragi"
)

func DownloadFile(dir, path, fileName string) (fullNewFilePath string) {
	log.Info("Downloading new version", "name", fileName, "path", path)
	// Get the data
	c := http.Client{}
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.WithError(err).Fatal("while creating new request")
	}
	if Creds != nil {
		r.SetBasicAuth(Creds.Username, Creds.Password)
	}
	resp, err := c.Do(r)
	if err != nil {
		log.WithError(err).Fatal("while executing download request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("there is no version", "os", runtime.GOOS, "arch", runtime.GOARCH)
	}

	fullNewFilePath = filepath.Clean(fmt.Sprintf("%s/%s", dir, fileName))
	out, err := os.OpenFile(fullNewFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.WithError(err).Fatal("while opening file to write download to")
	}
	defer out.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.WithError(err).Fatal("while copying downloaded body to file")
	}
	return
}
