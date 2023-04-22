package main

import (
	"flag"
	"fmt"
	"github.com/cantara/buri/maven"
	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/version"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"

	log "github.com/cantara/bragi"
	"github.com/cantara/buri/exec"
)

type program struct {
	path        string
	version     version.Version
	updatedTime time.Time
}

var displayHelpText bool
var repoUrl string
var groupId string
var artifactId string
var versionFilter string
var shouldRun bool
var numVersionsToKeep int
var kill bool
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
		packageTypeHelpText    = "Defines what type of service to work with (go, jar, tar, zip_jar)"
		killHelpText           = "kills the specified program rather than downloading and running it"
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
	flag.BoolVar(&kill, "kill", false, killHelpText)
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
		fmt.Print("All the following values is required:\n url to repo, groupId, artifactId, semantic version filter\n\n")
		flag.PrintDefaults()
		return
	} else if onlyKeepAlive && artifactId == "" {
		fmt.Print("ArtifactId is required when running with only keep alive active\n\n")
		flag.PrintDefaults()
		return
	}
	filter, err := version.ParseFilter(versionFilter)
	hd, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
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

	if shouldRun {
		os.Mkdir(hd+"/scripts", 0750)
		os.WriteFile(fmt.Sprintf("%s/scripts/restart_%s.sh", hd, linkName), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
~/scripts/kill_%[1]s.sh
sleep 5
~/scripts/start_%[1]s.sh
`, linkName)), 0750)
	}

	if strings.HasSuffix(packageType, "jar") {
		linkName = fmt.Sprintf("%s.jar", linkName)
	}

	command := []string{fmt.Sprintf("%s/%s", wd, linkName)}
	if strings.HasSuffix(packageType, "jar") {
		command = []string{"java", "-jar", command[0]}
	}
	if kill {
		exec.KillService(command)
		return
	}

	err = godotenv.Load(fmt.Sprintf(".env.buri.%s", strings.TrimSuffix(linkName, ".jar")))
	if err != nil {
		log.AddError(err).Info("while reading env for ", linkName)
	}
	foundNewerVersion := false
	defer func() {
		if !shouldRun && !onlyKeepAlive {
			return
		}
		if foundNewerVersion {
			exec.KillService(command)
		} else if exec.IsRunning(command) {
			return
		}
		exec.StartService(command, artifactId, linkName, wd)
		os.WriteFile(fmt.Sprintf("%s/scripts/start_%s.sh", hd, strings.TrimSuffix(linkName, ".jar")), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
%s > /dev/null
`, strings.Join(os.Args, " "))), 0750)
		os.WriteFile(fmt.Sprintf("%s/scripts/kill_%s.sh", hd, strings.TrimSuffix(linkName, ".jar")), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
%s -kill > /dev/null
`, strings.Join(os.Args, " "))), 0750)
	}()

	if onlyKeepAlive {
		return
	}
	url := fmt.Sprintf("%s/%s/%s", repoUrl, groupId, artifactId)

	var runningVersion version.Version
	/*
		if packageType == "go" {
			runningVersion = "v" + runningVersion
		}
	*/
	removeLink := false
	var versionsOnSystem []program
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		log.Debug("path: ", path)
		if path == linkName {
			linkPath, err := os.Readlink(path)
			if err != nil {
				log.AddError(err).Info("While trying to real symlink to get current verson")
				return nil
			}
			fileName := strings.ReplaceAll(strings.ReplaceAll(filepath.Base(linkPath), "-"+runtime.GOOS, ""), "-"+runtime.GOARCH, "")
			fileNameEls := strings.Split(fileName, "-")
			runningVersionString := fileNameEls[len(fileNameEls)-1]
			if strings.HasSuffix(packageType, "jar") { //should probably just do this always
				runningVersionString = strings.TrimSuffix(runningVersionString, ".jar")
			}
			runningVersion, err = version.ParseVersion(runningVersionString)
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
			name = strings.ReplaceAll(name, "-SNAPSHOT", "")
			versionString := strings.ReplaceAll(name, linkName+"-", "")
			if len(strings.Split(versionString, "-")) > 1 {
				log.Debug("skipping since it is probably a sub artifact: ", versionString)
				return nil
			}
			vers, err := version.ParseVersion(versionString)
			if err != nil {
				return err
			}
			versionsOnSystem = append(versionsOnSystem, program{
				path:    path,
				version: vers,
			})
		}
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	})
	log.Debug(versionsOnSystem)

	sort.Slice(versionsOnSystem, func(i, j int) bool {
		if filter.Snapshot {
			return false //snapshot.IsStrictlySemanticNewer(filter, versionsOnSystem[i].version, versionsOnSystem[j].version)
		}
		return version.IsStrictlySemanticNewer(filter, versionsOnSystem[i].version, versionsOnSystem[j].version)
	})
	defer func() {
		for i := 0; i < len(versionsOnSystem)-numVersionsToKeep; i++ {
			err = os.RemoveAll(versionsOnSystem[i].path)
			log.AddError(err).Info("while removing ", versionsOnSystem[i].path)
		}
	}()

	//log.Debug(url)
	params := maven.GetParamsURL("<td>(.+)</td>", url)
	//log.Debug(params)
	var programs []program
	for i := 1; i+1 < len(params); i++ {
		urlPars := maven.GetParams("<a href=\"(.+)\">(.+)</a>", params[i])
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
		//log.Println(urlPars)
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
		versionString := strings.TrimSuffix(urlPars[1], "/")
		vers, err := version.ParseVersion(versionString)
		if err != nil {
			log.AddError(err).Error("while parsing version")
			continue
		}
		if !filter.Matches(vers) {
			continue
		}

		log.Println(vers)
		programs = append(programs, program{
			path:    path,
			version: vers,
			//updatedTime: t,
		})
	}
	var newestP *program
	for i, p := range programs {
		if newestP == nil {
			newestP = &programs[i]
			continue
		}
		log.Debug("testing", "filter", filter, "v1", newestP.version, "v2", p.version)
		if version.IsStrictlySemanticNewer(filter, newestP.version, p.version) {
			log.Debug("Was newer")
			newestP = &programs[i]
		}
	}
	if newestP == nil {
		log.Fatal("No version found")
	}

	foundNewerVersion = runningVersion == version.Version{} || version.IsStrictlySemanticNewer(filter, runningVersion, newestP.version)

	if !foundNewerVersion {
		return
	}
	// Create the file

	var path string
	if len(subArtifact) == 1 {
		path = newestP.path
	} else {
		path = fmt.Sprintf("%s/%s/", strings.TrimSuffix(newestP.path, "/"), strings.Join(subArtifact[1:], "/"))
	}
	path = fmt.Sprintf("%s%s-%s", path, artifactId, newestP.version)
	fileName := fmt.Sprintf("%s-%s", strings.TrimSuffix(linkName, ".jar"), newestP.version)
	switch strings.ToLower(packageType) {
	case "go":
		fileName = fmt.Sprintf("%s-%s-%s", fileName, runtime.GOOS, runtime.GOARCH)
		path = fmt.Sprintf("%s-%s-%s", path, runtime.GOOS, runtime.GOARCH)
	case "jar":
		fileName = fmt.Sprintf("%s.jar", fileName)
		path = fmt.Sprintf("%s.jar", path)
	case "tgz":
		fileName = fmt.Sprintf("%s.tgz", fileName)
		path = fmt.Sprintf("%s.tgz", path)
	case "zip_jar":
		fileName = fmt.Sprintf("%s.zip", fileName)
		path = fmt.Sprintf("%s.zip", path)

	}
	maven.DownloadFile(path, fileName)
	if removeLink {
		err = os.Remove(linkName)
		if err != nil {
			log.Fatal(err)
		}
	}
	if packageType == "tgz" {
		pack.UnTGZ(fileName)
		os.Remove(fileName)
		fileName = strings.TrimSuffix(fileName, ".tgz")
		linkName = strings.TrimSuffix(linkName, ".tgz")
	} else if strings.HasPrefix(packageType, "zip") {
		linkName = strings.TrimSuffix(linkName, ".zip")
		pack.UnZip(fileName, linkName)
		os.Remove(fileName)
		fileName = strings.TrimSuffix(fileName, ".zip")
		linkName = strings.TrimSuffix(linkName, ".jar")
	}
	err = os.Symlink(fileName, linkName)
	if err != nil {
		log.Fatal(err)
	}
}
