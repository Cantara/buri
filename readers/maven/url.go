package maven

import (
	log "github.com/cantara/bragi/sbragi"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetFileNames(url string) (filenames []string) {
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
	filenames = parse(string(body))
	return
}

func parse(text string) (data []string) {
	tkn := html.NewTokenizer(strings.NewReader(text))
	var vals []string

	for {
		tt := tkn.Next()
		switch {
		case tt == html.ErrorToken:
			return vals
		case tt == html.StartTagToken:
			t := tkn.Token()
			if t.Data != "tr" {
				continue
			}
			for {
				tt = tkn.Next()
				if tt == html.ErrorToken {
					return vals
				}
				if tt == html.StartTagToken {
					t = tkn.Token()
					if t.Data != "th" {
						continue
					}
					break
				}
				if tt == html.EndTagToken {
					t = tkn.Token()
					if t.Data != "tr" {
						continue
					}
					break
				}
				if tt == html.TextToken {
					t := tkn.Token()
					d := strings.TrimSpace(t.Data)
					if d == "" {
						continue
					}
					vals = append(vals, d)
					break
				}
			}
		}
	}
}
