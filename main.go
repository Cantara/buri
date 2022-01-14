package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/cantara/bragi"
)

type program struct {
	url         string
	version     string
	updatedTime time.Time
}

func main() {
	fmt.Println("vim-go")
	repoUrl := "https://mvnrepo.cantara.no/content/repositories/releases"
	groupId := "no/cantara/vili"
	artifactId := "vili"
	url := fmt.Sprintf("%s/%s/%s", repoUrl, groupId, artifactId)

	params := getParamsURL("<td>(.+)</td>", url)
	var programs []program
	for i := 1; i+1 < len(params); i += 2 {
		urlPars := getParams("<a href=\"(.+)\">(.+)</a>", params[i])
		if len(urlPars) != 2 {
			log.Fatal("Wrong number of urls in path to version")
		}
		if !strings.HasSuffix(urlPars[1], "/") {
			continue
		}
		if !strings.HasPrefix(urlPars[1], "v") { //Could be removed if you don't want go specific selection
			continue
		}
		t, err := time.Parse("Mon Jan 02 15:04:05 MST 2006", params[i+1])
		if err != nil {
			log.Fatal(err)
		}
		programs = append(programs, program{
			url:         urlPars[0],
			version:     urlPars[1][:len(urlPars[1])-1],
			updatedTime: t,
		})
	}
	var newestP *program
	for _, p := range programs {
		if newestP == nil {
			newestP = &p
			continue
		}
		if isSemanticNewer("*.*.*", *newestP, p) {
			newestP = &p
		}
	}
	log.Println(newestP)
	// Create the file
	fileName := fmt.Sprintf("%s-%s", artifactId, newestP.version)
	out, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(fmt.Sprintf("%s%s", newestP.url, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err := os.Remove(artifactId)
	if err != nil {
		log.Fatal(err)
	}
	err := os.Symlink(fileName, artifactId)
	if err != nil {
		log.Fatal(err)
	}
}

func isSemanticNewer(filter string, p1, p2 program) bool {
	numLevels := 3
	levels := strings.Split(filter, ".")
	if len(levels) != numLevels {
		log.Fatal("Invalid semantic filter, expecting *.*.*")
	}
	p1v := strings.Split(p1.version[1:], ".")
	if len(p1v) != numLevels {
		log.Fatal("Invalid semantic version for arg 2, expecting v*.*.*")
	}
	p2v := strings.Split(p2.version[1:], ".")
	if len(p2v) != numLevels {
		log.Fatal("Invalid semantic version for arg 3, expecting v*.*.*")
	}
	for i := 0; i < numLevels; i++ {
		if levels[i] == "*" {
			v1, err := strconv.Atoi(p1v[i])
			if err != nil {
				log.Fatal(err)
			}
			v2, err := strconv.Atoi(p2v[i])
			if err != nil {
				log.Fatal(err)
			}
			if v1 < v2 {
				return true
			}
			if v1 > v2 {
				return false
			}
		}
	}
	return false
}

func getParamsURL(regEx, url string) (params []string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	params = getParams(regEx, string(body))
	return
}
func getParams(regEx, data string) (params []string) {
	var compRegEx = regexp.MustCompile(regEx)
	matches := compRegEx.FindAllStringSubmatch(data, -1)

	for _, line := range matches {
		for i, match := range line {
			//log.Println(match)
			if i == 0 {
				continue
			}
			params = append(params, match)
		}
	}
	return
}
