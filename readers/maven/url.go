package maven

import (
	"io"
	"net/http"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"golang.org/x/net/html"
)

// var Client = http.Client{}
var Creds *Credentials

type Credentials struct {
	Username string
	Password string
}

func GetFileNames(url string) (filenames []string) {
	log.Trace("getting filenames", "url", url)
	Client := http.Client{}
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithError(err).Fatal("while creating new request")
	}
	if Creds != nil {
		r.SetBasicAuth(Creds.Username, Creds.Password)
	}
	resp, err := Client.Do(r)
	if err != nil {
		log.WithError(err).Fatal("while executing request to get param urls")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Fatal("while reading body from request getting param urls")
	}
	log.Trace("maven", "body", body, "code", resp.StatusCode)
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
