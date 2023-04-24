package main

import (
	"flag"
	"fmt"
	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/readers/maven"
	versionFilter "github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic"
	"github.com/joho/godotenv"
	"os"
	"runtime"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/exec"
)

var displayHelpText bool
var repoUrl string
var groupId string
var artifactId string
var filterString string
var shouldRun bool
var numVersionsToKeep int
var kill bool
var onlyKeepAlive bool
var packageType string
var debug bool

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
		killHelpText           = "Kills the specified program rather than downloading and running it"
		debugHelpText          = "Enables debug and trace logging"
		helpText               = "Display help info"
	)

	flag.BoolVar(&displayHelpText, "h", false, helpText)
	flag.BoolVar(&displayHelpText, "help", false, helpText)
	flag.StringVar(&repoUrl, "u", "https://mvnrepo.cantara.no/content/repositories/releases", repoUrlHelpText)
	flag.StringVar(&groupId, "g", "", groupIdHelpText)
	flag.StringVar(&artifactId, "a", "", artifactIdHelpText)
	flag.StringVar(&filterString, "f", "*.*.*", versionFilterHelpText)
	flag.BoolVar(&shouldRun, "r", false, shouldRunHelpText)
	flag.IntVar(&numVersionsToKeep, "k", 4, versionsToKeepHelpText)
	flag.BoolVar(&kill, "kill", false, killHelpText)
	flag.BoolVar(&onlyKeepAlive, "o", false, onlyKeepAliveHelpText)
	flag.StringVar(&packageType, "t", "go", packageTypeHelpText)
	flag.BoolVar(&debug, "d", false, debugHelpText)
}

func main() {
	flag.Parse()
	if displayHelpText {
		flag.PrintDefaults()
		return
	}
	if !onlyKeepAlive && (repoUrl == "" || groupId == "" || artifactId == "" || filterString == "") {
		fmt.Print("All the following values is required:\n url to repo, groupId, artifactId, semantic version filter\n\n")
		flag.PrintDefaults()
		return
	} else if onlyKeepAlive && artifactId == "" {
		fmt.Print("ArtifactId is required when running with only keep alive active\n\n")
		flag.PrintDefaults()
		return
	}
	if debug {
		nl, err := log.NewDebugLogger()
		if err != nil {
			log.WithError(err).Fatal("while creating new debug logger")
		}
		nl.SetDefault()
	}
	filter, err := versionFilter.Parse(filterString)
	hd, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Fatal("while getting home dir")
	}
	groupId = strings.ReplaceAll(groupId, ".", "/")
	//log.SetLevel(log.DEBUG)

	subArtifact := strings.Split(artifactId, "/")
	artifactId = subArtifact[0]

	wd, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatal("while getting working dir")
	}

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
		log.WithError(err).Info("while reading env", "name", linkName)
	}
	foundNewerVersion := false
	defer func() {
		if !shouldRun && !onlyKeepAlive {
			return
		}
		os.WriteFile(fmt.Sprintf("%s/scripts/start_%s.sh", hd, strings.TrimSuffix(linkName, ".jar")), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
%s > /dev/null
`, strings.Join(os.Args, " "))), 0750)
		os.WriteFile(fmt.Sprintf("%s/scripts/kill_%s.sh", hd, strings.TrimSuffix(linkName, ".jar")), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
%s -kill > /dev/null
`, strings.Join(os.Args, " "))), 0750)
		if foundNewerVersion {
			exec.KillService(command)
		} else if exec.IsRunning(command) {
			return
		}
		exec.StartService(command, artifactId, linkName, wd)
	}()

	if onlyKeepAlive {
		return
	}
	/*
		var versionType Version[version.Base]
		versionType = reflect.TypeOf(version.Base)

		var runningVersion reflect.TypeOf("test")
			if packageType == "go" {
				runningVersion = "v" + runningVersion
			}
	*/

	diskFS := os.DirFS(wd)
	mavenPath, mavenVersion, removeLink, err := generic.NewestVersion(diskFS, filter, groupId, artifactId, linkName, packageType, repoUrl, numVersionsToKeep)
	if mavenVersion != "" {
		log.Info("new version found", "version", mavenVersion)
		foundNewerVersion = true
	}

	if !foundNewerVersion {
		log.Info("no new version found")
		return
	}
	// Create the file

	var path string
	if len(subArtifact) == 1 {
		path = mavenPath
	} else {
		path = fmt.Sprintf("%s/%s/", strings.TrimSuffix(mavenPath, "/"), strings.Join(subArtifact[1:], "/"))
	}
	path = fmt.Sprintf("%s%s-%s", path, artifactId, mavenVersion)
	fileName := fmt.Sprintf("%s-%s", strings.TrimSuffix(linkName, ".jar"), mavenVersion)
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
			log.WithError(err).Fatal("while removing link")
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
	os.Remove(linkName)
	err = os.Symlink(fileName, linkName)
	if err != nil {
		log.WithError(err).Fatal("while sym linking")
	}
}
