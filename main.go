package main

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/cantara/bragi"
)

type program struct {
	path        string
	version     string
	updatedTime time.Time
}

func main() {
	fmt.Println("vim-go")
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fileSystem := os.DirFS(wd)
	repoUrl := "https://mvnrepo.cantara.no/content/repositories/releases"
	groupId := "no/cantara/vili"
	artifactId := "vili"
	versionFilter := "*.*.*"
	numVersionsToKeep := 4
	url := fmt.Sprintf("%s/%s/%s", repoUrl, groupId, artifactId)
	runningVersion := ""
	removeLink := false
	var versionsOnSystem []program
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		log.Println(path)
		if d.IsDir() {
			return fs.SkipDir
		}
		if path == artifactId {
			linkPath, err := os.Readlink(path)
			if err != nil {
				log.Fatal(err)
			}
			linkPathEls := strings.Split(linkPath, "/")
			fileNameEls := strings.Split(linkPathEls[len(linkPathEls)-1], "-")
			runningVersion = fileNameEls[len(fileNameEls)-1]
			removeLink = true
			return nil
		}
		if strings.HasPrefix(path, artifactId+"-v") { //TODO: Should be a regex so it can handle programs who start with artifactId and then continues with -vSomething
			linkPathEls := strings.Split(path, "/")
			fileNameEls := strings.Split(linkPathEls[len(linkPathEls)-1], "-")
			versionsOnSystem = append(versionsOnSystem, program{
				path:    path,
				version: fileNameEls[len(fileNameEls)-1],
			})
		}
		return nil
	})
	sort.Slice(versionsOnSystem, func(i, j int) bool {
		return isSemanticNewer("*.*.*", versionsOnSystem[i].version, versionsOnSystem[j].version) // Could also be dependent on our semantic version tactic
	})
	defer func() {
		for i := 0; i < len(versionsOnSystem)-numVersionsToKeep; i++ {
			os.Remove(versionsOnSystem[i].path)
		}
	}()

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
			path:        urlPars[0],
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
		if isSemanticNewer(versionFilter, newestP.version, p.version) {
			newestP = &p
		}
	}

	stdOut, err := os.OpenFile(fmt.Sprintf("%s/%sOut", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	stdErr, err := os.OpenFile(fmt.Sprintf("%s/%sErr", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		cmd := exec.Command("ps", "h", "-C", artifactId)
		out, err := cmd.Output()
		if err != nil {
			log.Println(err)
		}
		if len(out) != 0 {
			log.Println(string(out), err)
			return
		}

		cmd = exec.Command(fmt.Sprintf("%s/%s", wd, artifactId))
		cmd.Stdout = stdOut
		cmd.Stderr = stdErr
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if !(runningVersion == "" || isSemanticNewer(versionFilter, runningVersion, newestP.version)) {
		return
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
	resp, err := http.Get(fmt.Sprintf("%s%s", newestP.path, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if removeLink {
		err = os.Remove(artifactId)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = os.Symlink(fileName, artifactId)
	if err != nil {
		log.Fatal(err)
	}
}

func isSemanticNewer(filter string, p1, p2 string) bool {
	numLevels := 3
	levels := strings.Split(filter, ".")
	if len(levels) != numLevels {
		log.Fatal("Invalid semantic filter, expecting *.*.*")
	}
	p1v := strings.Split(p1[1:], ".")
	if len(p1v) != numLevels {
		log.Fatal("Invalid semantic version for arg 2, expecting v*.*.*")
	}
	p2v := strings.Split(p2[1:], ".")
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
