package cmd

import (
	"fmt"
	"io/fs"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/download"
	"github.com/cantara/buri/readers/maven"
	"github.com/cantara/buri/run"
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

func fixArtifactStrings(groupIdIn, artifactIdIn, packageType string) (groupId, artifactId, artifactName, linkName string, subArtifact []string) {
	groupId = strings.ReplaceAll(groupIdIn, ".", "/")
	artifactId = strings.ReplaceAll(artifactIdIn, ".", "/")
	subArtifact = strings.Split(artifactId, "/")
	artifactId = subArtifact[0]
	artifactName = strings.Join(subArtifact, "-")

	linkName = artifactId
	if len(subArtifact) > 1 {
		linkName = fmt.Sprintf("%s-%s", linkName, strings.Join(subArtifact[1:], "-"))
	}

	if strings.HasSuffix(packageType, "jar") {
		linkName = fmt.Sprintf("%s.jar", linkName)
	}
	return
}

type PackageType string

const (
	PackageJar = PackageType("jar")
	PackageGo  = PackageType("go")
	PackageTar = PackageType("tar")
	PackageZip = PackageType("zip")
)

func (s *PackageType) String() string {
	return fmt.Sprint(*s)
}

func serviceTypeFromString(s string) (pt PackageType) {
	switch strings.ToLower(s) {
	case "java":
		pt = PackageJar
	case "jar":
		pt = PackageJar
	case "go":
		pt = PackageGo
	case "tar":
		pt = PackageTar
	case "zip":
		pt = PackageZip
	case "raw_zip":
		pt = PackageZip
	default:
		//err = errors.New("unsuported service type")
		log.Info("service type not found. treating as website / frontend") //Could be smart to return to error and use tag website and artifact for name of website
		pt = PackageType(fmt.Sprintf("website_%s", s))
	}
	return
}

type PackageRepo struct {
}

func (pr PackageRepo) DownloadFile(dir, path, filename string) string {
	return maven.DownloadFile(dir, path, filename)
}

func (pr PackageRepo) NewestVersion(localFS fs.FS, f filter.Filter, groupId, artifactId, linkName, packageType, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error) {
	return generic.NewestVersion(localFS, f, groupId, artifactId, linkName, packageType, repoUrl, numVersionsToKeep)
}

type ArtifactDownloader interface {
	Download(localFS fs.FS, pr download.PackageRepo, packageType, linkName, artifactId, groupId, repoUrl string, subArtifact []string, f filter.Filter) (newFileName string)
}

type ConfigHandler interface {
	Config(artifactName string) (repoUrl string, f filter.Filter)
}

type Runner interface {
	IsRunning(artifactId string) bool
	Start(dir, artifactId, name, linkName, packageType string, foundNewerVersion bool)
}

type LocalRunner struct {
}

func (_ LocalRunner) Start(dir, artifactId, name, linkName, packageType string, foundNewerVersion bool) {
	run.Run(dir, artifactId, name, linkName, packageType, foundNewerVersion)
}

func (_ LocalRunner) IsRunning(artifactId string) bool {
	return false
}
