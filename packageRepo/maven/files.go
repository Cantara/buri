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

func GetTableValues(url string) (values []string) {
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
	values = parse(string(body))
	return
}

func parse(text string) []string {
	tkn := html.NewTokenizer(strings.NewReader(text))
	var vals []string

	for {
		tt := tkn.Next()
		switch {
		case tt == html.ErrorToken:
			return uniqueify(vals)
		case tt == html.StartTagToken:
			t := tkn.Token()
			if t.Data != "tr" {
				continue
			}
			for {
				tt = tkn.Next()
				if tt == html.ErrorToken {
					return uniqueify(vals)
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
					if d == "Parent Directory" {
						continue
					}
					d = strings.TrimSuffix(d, "/")
					vals = append(vals, d)
					break
				}
			}
		}
	}
}

func uniqueify(in []string) []string {
	out := make([]string, len(in))
	numEl := 0
	for _, el := range in {
		if contains(out[:numEl], el) {
			continue
		}
		out[numEl] = el
		numEl++
	}
	return out[:numEl]
}

func contains(arr []string, el string) bool {
	for _, e := range arr {
		if el == e {
			return true
		}
	}
	return false
}
