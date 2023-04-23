package maven

import (
	log "github.com/cantara/bragi/sbragi"
	"io"
	"net/http"
	"os"
	"runtime"
)

func DownloadFile(path, fileName string) {
	log.Info("Downloading new version", "name", fileName, "path", path)
	// Get the data
	c := http.Client{}
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.WithError(err).Fatal("while creating new request")
	}
	if os.Getenv("username") != "" {
		r.SetBasicAuth(os.Getenv("username"), os.Getenv("password"))
	}
	resp, err := c.Do(r)
	if err != nil {
		log.WithError(err).Fatal("while executing download request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("there is no version", "os", runtime.GOOS, "arch", runtime.GOARCH)
	}

	out, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.WithError(err).Fatal("while opening file to write download to")
	}
	defer out.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.WithError(err).Fatal("while copying downloaded body to file")
	}
}
