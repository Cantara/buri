package maven

import (
	log "github.com/cantara/bragi/sbragi"
	"io"
	"net/http"
	"os"
	"regexp"
)

func GetParamsURL(regEx, url string) (params []string) {
	c := http.Client{}
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithError(err).Fatal("while creating new request")
	}
	if os.Getenv("username") != "" {
		r.SetBasicAuth(os.Getenv("username"), os.Getenv("password"))
	}
	resp, err := c.Do(r)
	if err != nil {
		log.WithError(err).Fatal("while executing request to get param urls")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Fatal("while reading body from request getting param urls")
	}
	params = GetParams(regEx, string(body))
	log.Debug("maven", "params", params)
	return
}

func GetParams(regEx, data string) (params []string) {
	var compRegEx = regexp.MustCompile(regEx)
	matches := compRegEx.FindAllStringSubmatch(data, -1)

	for _, line := range matches {
		for i, match := range line {
			//log.Debug(match)
			if i == 0 {
				continue
			}
			params = append(params, match)
		}
	}
	return
}
