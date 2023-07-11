/*
Copyright © 2023 Sindre Brurberg

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/download"
	"github.com/cantara/buri/readers/maven"
	"github.com/cantara/buri/version/filter"
	"github.com/cantara/buri/version/generic"
	"github.com/cantara/buri/version/snapshot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var filterFlagVar = filterFlag{}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <packageType>",
	Short: "Downloads version of software that matches filter",
	Long: `Uses the filter provided to download the specified software from Nexus.

The software will be downloaded to the working directory and unpackaged if needed.`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{string(PackageJar), string(PackageGo), string(PackageTar)}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		packageType := args[0]
		artifactId, _ := cmd.Flags().GetString("artifact")
		groupId, _ := cmd.Flags().GetString("group")

		groupId, artifactId, artifactName, linkName, subArtifact := fixArtifactStrings(groupId, artifactId, packageType)
		repoUrl, f := getConfig(artifactName)

		wd, err := os.Getwd()
		if err != nil {
			log.WithError(err).Fatal("while getting working dir")
		}

		download.Download(os.DirFS(wd), PackageRepo{}, packageType, linkName, artifactId, groupId, repoUrl, subArtifact, f)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().VarP(&filterFlagVar, "filter", "f", "Filter for matching versions")
	getCmd.Flags().StringP("artifact", "a", "buri", "Artifact id of the software")
	getCmd.Flags().StringP("group", "g", "no.cantara.gotools", "Artifact group of the software")

	getCmd.RegisterFlagCompletionFunc("filter", filterCompletion)
}

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

func getConfig(artifactName string) (repoUrl string, f filter.Filter) {
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

func (pr PackageRepo) DownloadFile(dir, path, filename string) {
	maven.DownloadFile(dir, path, filename)
}

func (pr PackageRepo) NewestVersion(localFS fs.FS, f filter.Filter, groupId, artifactId, linkName, packageType, repoUrl string, numVersionsToKeep int) (mavenPath, mavenVersion string, removeLink bool, err error) {
	return generic.NewestVersion(localFS, f, groupId, artifactId, linkName, packageType, repoUrl, numVersionsToKeep)
}
