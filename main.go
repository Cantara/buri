package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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

const (
	repoUrlHelpText        = "nexus `repo` url\nex: https://mvnrepo.cantara.no/content/repositories/releases"
	groupIdHelpText        = "maven `group` id"
	artifactIdHelpText     = "maven `artifact` id"
	versionFilterHelpText  = "Semantic version `filter`"
	shouldRunHelpText      = "Should execute the downloaded program if it is not running or if it is downloaded"
	versionsToKeepHelpText = "Number of `versions` to keep in directory as backup"
	onlyKeepAliveHelpText  = "Makes it so buri only keeps the program running"
	helpText               = "Display help info"
)

func main() {
	fmt.Println("vim-go")
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fileSystem := os.DirFS(wd)
	var displayHelpText bool
	flag.BoolVar(&displayHelpText, "h", false, helpText)
	var repoUrl string
	flag.StringVar(&repoUrl, "u", "https://mvnrepo.cantara.no/content/repositories/releases", repoUrlHelpText)
	var groupId string
	flag.StringVar(&groupId, "g", "", groupIdHelpText)
	var artifactId string
	flag.StringVar(&artifactId, "a", "", artifactIdHelpText)
	var versionFilter string
	flag.StringVar(&versionFilter, "f", "*.*.*", versionFilterHelpText)
	var shouldRun bool
	flag.BoolVar(&shouldRun, "r", false, shouldRunHelpText)
	var numVersionsToKeep int
	flag.IntVar(&numVersionsToKeep, "k", 4, versionsToKeepHelpText)
	var onlyKeepAlive bool
	flag.BoolVar(&onlyKeepAlive, "o", false, onlyKeepAliveHelpText)
	flag.Parse()
	if displayHelpText {
		flag.PrintDefaults()
		return
	}
	if !onlyKeepAlive && (repoUrl == "" || groupId == "" || groupId == "" || artifactId == "" || versionFilter == "") {
		fmt.Println("All the following values is required:\n url to repo, groupId, artifactId, semantic version filter\n")
		flag.PrintDefaults()
		return
	} else if onlyKeepAlive && artifactId == "" {
		fmt.Println("ArtifactId is required when running with only keep alive active\n")
		flag.PrintDefaults()
		return
	}

	foundNewerVersion := false
	defer func() {
		if !shouldRun && !onlyKeepAlive {
			return
		}
		ex := fmt.Sprintf("%s/%s", wd, artifactId)
		cmd := exec.Command("pgrep", "-u", strconv.Itoa(os.Getuid()), artifactId) //Breaking compatibility with windows / probably
		out, err := cmd.Output()
		if err != nil {
			log.Println(err)
		}
		if !foundNewerVersion {
			//cmd := exec.Command("ps", "h", "-C", ex)
			if len(out) != 0 {
				log.Println(string(out), err)
				return
			}
		}
		err = exec.Command("pkill", "-9", "-P", strings.ReplaceAll(string(out), "\n", ",")).Run()
		if err != nil {
			log.Println(err)
		}
		err = exec.Command("pkill", "-9", artifactId).Run()
		if err != nil {
			log.Println(err)
		}

		stdOut, err := os.OpenFile(fmt.Sprintf("%s/%sOut", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		stdErr, err := os.OpenFile(fmt.Sprintf("%s/%sErr", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		cmd = exec.Command(ex)
		cmd.Stdout = stdOut
		cmd.Stderr = stdErr
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if onlyKeepAlive {
		return
	}
	url := fmt.Sprintf("%s/%s/%s", repoUrl, groupId, artifactId)

	runningVersion := "v0.0.0"
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
				log.AddError(err).Info("While trying to real symlink to get current verson")
				return nil
			}
			fileName := strings.ReplaceAll(strings.ReplaceAll(filepath.Base(linkPath), "-"+runtime.GOOS, ""), "-"+runtime.GOARCH, "")
			fileNameEls := strings.Split(fileName, "-")
			runningVersion = fileNameEls[len(fileNameEls)-1]
			removeLink = true
			return nil
		}
		if strings.HasPrefix(path, artifactId+"-") {
			name := filepath.Base(path)
			name = strings.ReplaceAll(name, "-"+runtime.GOOS, "")
			name = strings.ReplaceAll(name, "-"+runtime.GOARCH, "")
			versionsOnSystem = append(versionsOnSystem, program{
				path:    path,
				version: strings.ReplaceAll(name, artifactId+"-", ""),
			})
		}
		return nil
	})
	fmt.Println("HITT")
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
		fmt.Println(urlPars)
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
	for i, p := range programs {
		if newestP == nil {
			newestP = &programs[i]
			continue
		}
		if isSemanticNewer(versionFilter, newestP.version, p.version) {
			log.Println("Was newer")
			newestP = &programs[i]
		}
	}

	foundNewerVersion = runningVersion == "" || isSemanticNewer(versionFilter, runningVersion, newestP.version)

	if !foundNewerVersion {
		return
	}
	log.Println(newestP)
	// Create the file
	fileName := fmt.Sprintf("%s-%s-%s-%s", artifactId, newestP.version, runtime.GOOS, runtime.GOARCH)
	fmt.Println(newestP)
	downloadFile(newestP.path, fileName)
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

func downloadFile(path, fileName string) {

	// Get the data
	resp, err := http.Get(fmt.Sprintf("%s%s", path, fileName))
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

func isSemanticNewer(filter string, p1, p2 string) bool {
	log.Printf("Testing %s vs %s with filter %s\n", p1, p2, filter)
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
