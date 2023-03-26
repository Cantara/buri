package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"io"
	"io/fs"
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

	"github.com/joho/godotenv"

	log "github.com/cantara/bragi"
)

type program struct {
	path        string
	version     string
	updatedTime time.Time
}

var displayHelpText bool
var repoUrl string
var groupId string
var artifactId string
var versionFilter string
var shouldRun bool
var numVersionsToKeep int
var onlyKeepAlive bool
var packageType string

func init() {
	const (
		repoUrlHelpText        = "nexus `repo` url\nex: https://mvnrepo.cantara.no/content/repositories/releases"
		groupIdHelpText        = "maven `group` id"
		artifactIdHelpText     = "maven `artifact` id"
		versionFilterHelpText  = "Semantic version `filter`"
		shouldRunHelpText      = "Should execute the downloaded program if it is not running or if it is downloaded"
		versionsToKeepHelpText = "Number of `versions` to keep in directory as backup"
		onlyKeepAliveHelpText  = "Makes it so buri only keeps the program running"
		packageTypeHelpText    = "Defines what type of service to work with"
		helpText               = "Display help info"
	)

	flag.BoolVar(&displayHelpText, "h", false, helpText)
	flag.BoolVar(&displayHelpText, "help", false, helpText)
	flag.StringVar(&repoUrl, "u", "https://mvnrepo.cantara.no/content/repositories/releases", repoUrlHelpText)
	flag.StringVar(&groupId, "g", "", groupIdHelpText)
	flag.StringVar(&artifactId, "a", "", artifactIdHelpText)
	flag.StringVar(&versionFilter, "f", "*.*.*", versionFilterHelpText)
	flag.BoolVar(&shouldRun, "r", false, shouldRunHelpText)
	flag.IntVar(&numVersionsToKeep, "k", 4, versionsToKeepHelpText)
	flag.BoolVar(&onlyKeepAlive, "o", false, onlyKeepAliveHelpText)
	flag.StringVar(&packageType, "t", "go", packageTypeHelpText)
}

func main() {
	flag.Parse()
	if displayHelpText {
		flag.PrintDefaults()
		return
	}
	if !onlyKeepAlive && (repoUrl == "" || groupId == "" || artifactId == "" || versionFilter == "") {
		fmt.Println("All the following values is required:\n url to repo, groupId, artifactId, semantic version filter\n")
		flag.PrintDefaults()
		return
	} else if onlyKeepAlive && artifactId == "" {
		fmt.Println("ArtifactId is required when running with only keep alive active\n")
		flag.PrintDefaults()
		return
	}
	groupId = strings.ReplaceAll(groupId, ".", "/")
	log.SetLevel(log.DEBUG)

	subArtifact := strings.Split(artifactId, "/")
	artifactId = subArtifact[0]

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fileSystem := os.DirFS(wd)

	linkName := artifactId
	if len(subArtifact) > 1 {
		linkName = fmt.Sprintf("%s-%s", linkName, strings.Join(subArtifact[1:], "-"))
	}

	err = godotenv.Load(fmt.Sprintf(".env.buri.%s", linkName))
	if err != nil {
		log.AddError(err).Info("while reading env for ", linkName)
	}

	if packageType == "jar" {
		linkName = fmt.Sprintf("%s.jar", linkName)
	}
	foundNewerVersion := false
	defer func() {
		if !shouldRun && !onlyKeepAlive {
			return
		}
		command := []string{fmt.Sprintf("%s/%s", wd, linkName)}
		if packageType == "jar" {
			command = []string{"java", "-jar", command[0]}
		}
		commandString := strings.Join(command, " ")

		procs, err := process.Processes()
		if err != nil {
			log.AddError(err).Fatal("while getting processes")
		}
		for _, proc := range procs {
			if uids, err := proc.Uids(); err != nil || int(uids[0]) != os.Getuid() {
				continue
			}
			cmd, err := proc.Cmdline()
			if err != nil {
				log.AddError(err).Warning("while getting cmd")
				continue
			}
			if cmd != commandString {
				continue
			}
			log.Info(cmd)
			if !foundNewerVersion {
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			err = proc.TerminateWithContext(ctx)
			cancel()
			if err != nil {
				err = proc.Kill()
				if err != nil {
					log.AddError(err).Error("while terminating service", "cmd", cmd)
				}
			}
			break
		}

		stdOut, err := os.OpenFile(fmt.Sprintf("%s/%sOut", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		stdErr, err := os.OpenFile(fmt.Sprintf("%s/%sErr", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		var cmd *exec.Cmd
		if len(command) == 1 {
			cmd = exec.Command(command[0])
		} else {
			cmd = exec.Command(command[0], command[1:]...)
		}
		var envMap map[string]string
		envMap, err = godotenv.Read(".env."+strings.TrimSuffix(linkName, ".jar"), strings.TrimSuffix(linkName, ".jar")+".env")
		if err != nil {
			log.AddError(err).Info("while reading env files")
		}
		env := make([]string, len(envMap))
		i := 0
		for k, v := range envMap {
			env[i] = fmt.Sprintf("%s=%s", k, v)
			i++
		}
		cmd.Env = append(cmd.Environ(), env...)
		cmd.Stdout = stdOut
		cmd.Stderr = stdErr
		log.Debug(cmd)
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	if onlyKeepAlive {
		return
	}
	url := fmt.Sprintf("%s/%s/%s", repoUrl, groupId, artifactId)

	runningVersion := "0.0.0"
	if packageType == "go" {
		runningVersion = "v" + runningVersion
	}
	removeLink := false
	var versionsOnSystem []program
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		log.Debug(path)
		if d.IsDir() {
			return fs.SkipDir
		}
		if path == linkName {
			linkPath, err := os.Readlink(path)
			if err != nil {
				log.AddError(err).Info("While trying to real symlink to get current verson")
				return nil
			}
			fileName := strings.ReplaceAll(strings.ReplaceAll(filepath.Base(linkPath), "-"+runtime.GOOS, ""), "-"+runtime.GOARCH, "")
			fileNameEls := strings.Split(fileName, "-")
			runningVersion = fileNameEls[len(fileNameEls)-1]
			if packageType == "jar" {
				runningVersion = strings.TrimSuffix(runningVersion, ".jar")
			}
			removeLink = true
			return nil
		}
		if strings.HasPrefix(path, linkName+"-") {
			log.Debug(path)
			log.Debug(linkName)
			name := filepath.Base(path)
			name = strings.TrimSuffix(name, ".jar")
			name = strings.ReplaceAll(name, "-"+runtime.GOOS, "")
			name = strings.ReplaceAll(name, "-"+runtime.GOARCH, "")
			vers := strings.ReplaceAll(name, linkName+"-", "")
			if len(strings.Split(vers, "-")) > 1 {
				log.Debug("skipping since it is probably a sub artifact: ", vers)
				return nil
			}
			versionsOnSystem = append(versionsOnSystem, program{
				path:    path,
				version: vers,
			})
		}
		return nil
	})
	log.Debug(versionsOnSystem)

	sort.Slice(versionsOnSystem, func(i, j int) bool {
		return isSemanticNewer("*.*.*", versionsOnSystem[i].version, versionsOnSystem[j].version) // Could also be dependent on our semantic version tactic
	})
	defer func() {
		for i := 0; i < len(versionsOnSystem)-numVersionsToKeep; i++ {
			err = os.RemoveAll(versionsOnSystem[i].path)
			log.AddError(err).Info("while removing ", versionsOnSystem[i].path)
		}
	}()

	log.Debug(url)
	params := getParamsURL("<td>(.+)</td>", url)
	log.Debug(params)
	var programs []program
	for i := 1; i+1 < len(params); i++ {
		urlPars := getParams("<a href=\"(.+)\">(.+)</a>", params[i])
		if len(urlPars) != 2 {
			//log.Fatal("Wrong number of urls in path to version")
			continue
		}
		if !strings.HasSuffix(urlPars[0], "/") {
			continue
		}
		if packageType == "go" && !strings.HasPrefix(urlPars[1], "v") { //Could be removed if you don't want go specific selection
			continue
		}
		log.Println(urlPars)
		/*
			t, err := time.Parse("Mon Jan 02 15:04:05 MST 2006", params[i+1])
			if err != nil {
				log.Fatal(err)
			}
		*/
		var path string
		if strings.HasPrefix(urlPars[0], "http") {
			path = urlPars[0]
		} else {
			path = fmt.Sprintf("%s/%s", strings.TrimSuffix(strings.ReplaceAll(url, "service/rest/repository/browse/", "repository/"),
				"/"), strings.TrimPrefix(urlPars[0], "/"))
		}
		programs = append(programs, program{
			path:    path,
			version: strings.TrimSuffix(urlPars[1], "/"),
			//updatedTime: t,
		})
		log.Println(programs)
	}
	var newestP *program
	for i, p := range programs {
		if newestP == nil {
			newestP = &programs[i]
			continue
		}
		if isSemanticNewer(versionFilter, newestP.version, p.version) {
			log.Debug("Was newer")
			newestP = &programs[i]
		}
	}
	if newestP == nil {
		log.Fatal("No version found")
	}

	foundNewerVersion = runningVersion == "" || isSemanticNewer(versionFilter, runningVersion, newestP.version)

	if !foundNewerVersion {
		return
	}
	// Create the file

	fileName := fmt.Sprintf("%s-%s", strings.TrimSuffix(linkName, ".jar"), newestP.version)
	if packageType == "go" {
		fileName = fmt.Sprintf("%s-%s-%s", fileName, runtime.GOOS, runtime.GOARCH)
	}
	if packageType == "jar" {
		fileName = fmt.Sprintf("%s.jar", fileName)
	}
	if packageType == "tgz" {
		fileName = fmt.Sprintf("%s.tgz", fileName)
	}
	var path string
	if len(subArtifact) == 1 {
		path = newestP.path
	} else {
		path = fmt.Sprintf("%s/%s/", strings.TrimSuffix(newestP.path, "/"), strings.Join(subArtifact[1:], "/"))
	}
	path = fmt.Sprintf("%s%s-%s", path, artifactId, newestP.version)
	if packageType == "go" {
		path = fmt.Sprintf("%s-%s-%s", path, runtime.GOOS, runtime.GOARCH)
	}
	if packageType == "jar" {
		path = fmt.Sprintf("%s.jar", path)
	}
	if packageType == "tgz" {
		path = fmt.Sprintf("%s.tgz", path)
	}
	downloadFile(path, fileName)
	if removeLink {
		err = os.Remove(linkName)
		if err != nil {
			log.Fatal(err)
		}
	}
	if packageType == "tgz" {
		unpackTGZ(fileName)
		os.Remove(fileName)
		fileName = strings.TrimSuffix(fileName, ".tgz")
		linkName = strings.TrimSuffix(linkName, ".tgz")
	}
	err = os.Symlink(fileName, linkName)
	if err != nil {
		log.Fatal(err)
	}
}

func downloadFile(path, fileName string) {
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

func isSemanticNewer(filter string, p1, p2 string) bool {
	log.Printf("Testing %s vs %s with filter %s\n", p1, p2, filter)
	if packageType == "go" {
		p1 = p1[1:]
		p2 = p2[1:]
	}
	levels := strings.Split(filter, ".")
	if len(levels) != 3 {
		log.Fatal("Invalid semantic filter, expecting *.*.*")
	}
	p1v := strings.Split(p1, ".")
	if len(p1v) == 1 {
		p1v = append(p1v, []string{"0", "0"}...)
	} else if len(p1v) == 2 {
		p1v = append(p1v, "0")
	} else if len(p1v) > 3 {
		log.Fatal("Invalid semantic version for arg 2, expecting v*.*.* ", p1v)
	}
	p2v := strings.Split(p2, ".")
	if len(p2v) == 1 {
		p2v = append(p2v, []string{"0", "0"}...)
	} else if len(p2v) == 2 {
		p2v = append(p2v, "0")
	} else if len(p2v) > 3 {
		log.Fatal("Invalid semantic version for arg 3, expecting v*.*.*")
	}
	for i := 0; i < 3; i++ {
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
	params = getParams(regEx, string(body))
	return
}
func getParams(regEx, data string) (params []string) {
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

func unpackTGZ(srcFile string) (err error) {
	base := strings.TrimSuffix(srcFile, ".tgz")
	os.Mkdir(base, 0750)
	tgz, err := os.Open(srcFile)
	if err != nil {
		return
	}
	defer tgz.Close()

	gzf, err := gzip.NewReader(tgz)
	if err != nil {
		return
	}

	tarReader := tar.NewReader(gzf)
	for {
		header, err := tarReader.Next()
		if err != nil {
			return err
		}

		name := header.Name
		fmt.Println(name)

		switch header.Typeflag {
		case tar.TypeDir:
			fmt.Println("Directory:", name)
			os.Mkdir(fmt.Sprintf("%s/%s", base, name), 0750)
		case tar.TypeReg:
			fmt.Println("Regular file:", name)
			func() {
				fn := fmt.Sprintf("%s/%s", base, name)
				f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
				if err != nil {
					log.AddError(err).Error("while opening file, ", name)
					return
				}
				defer f.Close()
				_, err = io.Copy(f, tarReader)
				if err != nil {
					log.AddError(err).Error("while reading file ", name)
					return
				}
			}()
		default:
			log.Warning("not a known file type, ", name, ", ", header.Typeflag)
		}
	}
}
