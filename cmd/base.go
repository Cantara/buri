package cmd

import (
	"fmt"
	"io/fs"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/download"
	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/packageRepo/maven"
	"github.com/cantara/buri/runner/start"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic"
	"github.com/cantara/buri/version/snapshot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type filterFlag struct {
	*filter.Filter
}

func (f *filterFlag) String() string {
	if f.Filter == nil {
		return "*.*.*"
	}
	return f.Filter.String()
}

func (f *filterFlag) Set(s string) error {
	ft, err := filter.Parse(s)
	if err != nil {
		return err
	}
	*f = filterFlag{&ft}
	return nil
}

func (f *filterFlag) Type() string {
	return "filter"
}

func filterCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"*.*.*\tall releases",
		"*.*.*-SNAPSHOT\tall snapshots",
	}, cobra.ShellCompDirectiveDefault
}

type ViperConfigHandler struct{}

func (_ ViperConfigHandler) Config(artifactName string) (repoUrl string, f filter.Filter) {
	releaseUrl := viper.GetString("release_url")
	snapshotUrl := viper.GetString("snapshot_url")
	username := viper.GetString("username")
	password := viper.GetString("password")
	artifactConfig := viper.GetStringMapString(artifactName)

	if _, ok := artifactConfig["release_url"]; ok {
		releaseUrl = artifactConfig["release_url"]
	}
	if _, ok := artifactConfig["snapshot_url"]; ok {
		snapshotUrl = artifactConfig["snapshot_url"]
	}
	if _, ok := artifactConfig["username"]; ok {
		username = artifactConfig["username"]
	}
	if _, ok := artifactConfig["password"]; ok {
		password = artifactConfig["password"]
	}

	f = filter.AllReleases
	if filterFlagVar.Filter == nil {
		filterString, ok := artifactConfig["filter"]
		if !ok {
			filterString = viper.GetString("filter")
		}
		if filterString != "" {
			var err error
			f, err = filter.Parse(filterString)
			if err != nil {
				log.WithError(err).Fatal("while parsing filter string from config")
			}
		}
	} else {
		f = *filterFlagVar.Filter
	}

	repoUrl = releaseUrl
	switch f.Type {
	case snapshot.Type:
		repoUrl = snapshotUrl
	}
	if repoUrl == "" {
		repoUrl = "https://mvnrepo.cantara.no/content/repositories/releases"
	}

	if username != "" {
		maven.Creds = &maven.Credentials{
			Username: username,
			Password: password,
		}
	}
	return
}

func fixArtifactStrings(groupIdIn, artifactIdIn string, packageType pack.Type) (groupId, artifactId, artifactName, linkName string, subArtifact []string) {
	groupId = strings.ReplaceAll(groupIdIn, ".", "/")
	artifactId = strings.ReplaceAll(artifactIdIn, ".", "/")
	subArtifact = strings.Split(artifactId, "/")
	artifactId = subArtifact[0]
	artifactName = strings.Join(subArtifact, "-")

	linkName = artifactId
	if len(subArtifact) > 1 {
		linkName = fmt.Sprintf("%s-%s", linkName, strings.Join(subArtifact[1:], "-"))
	}

	switch packageType {
	case pack.Jar:
		linkName = fmt.Sprintf("%s.jar", linkName)
	case pack.Tar:
		linkName = fmt.Sprintf("%s.tgz", linkName)
	case pack.Zip:
		linkName = fmt.Sprintf("%s.zip", linkName)
	}
	return
}

type PackageRepo struct {
}

func (pr PackageRepo) DownloadFile(dir, path, filename string) string {
	return maven.DownloadFile(dir, path, filename)
}

func (pr PackageRepo) NewestVersion(localFS fs.FS, f filter.Filter, groupId, artifactId, linkName string, packageType pack.Type, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error) {
	return generic.NewestVersion(localFS, f, groupId, artifactId, linkName, packageType, repoUrl, numVersionsToKeep)
}

type ArtifactDownloader interface {
	Download(localFS fs.FS, pr download.PackageRepo, packageType pack.Type, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (newFileName string)
}

type Unpacker interface {
	Unpack(localFS fs.FS, fullFileName, linkName string, packageType pack.Type)
}

type ConfigHandler interface {
	Config(artifactName string) (repoUrl string, f filter.Filter)
}

type Runner interface {
	IsRunning(artifactId string) bool
	Start(dir, artifactId, name, linkName string, packageType pack.Type, foundNewerVersion bool)
}

type LocalRunner struct {
}

func (_ LocalRunner) Start(dir, artifactId, name, linkName string, packageType pack.Type, foundNewerVersion bool) {
	start.Run(dir, artifactId, name, linkName, packageType, foundNewerVersion)
}

func (_ LocalRunner) IsRunning(artifactId string) bool {
	return false
}
