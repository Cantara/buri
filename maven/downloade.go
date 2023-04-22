package maven

import (
	log "github.com/cantara/bragi"
	"io"
	"net/http"
	"os"
	"runtime"
)

func DownloadFile(path, fileName string) {
	log.Info("Downloading new version, ", fileName, "\n", path)
	// Get the data
	c := http.Client{}
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.AddError(err).Fatal("while creating new request")
	}
	if os.Getenv("username") != "" {
		r.SetBasicAuth(os.Getenv("username"), os.Getenv("password"))
	}
	resp, err := c.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("there is no version for ", runtime.GOOS, " ", runtime.GOARCH)
	}

	out, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
