package maven

import (
	log "github.com/cantara/bragi"
	"io"
	"net/http"
	"os"
	"regexp"
)

func GetParamsURL(regEx, url string) (params []string) {
	c := http.Client{}
	r, err := http.NewRequest("GET", url, nil)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	params = GetParams(regEx, string(body))
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
