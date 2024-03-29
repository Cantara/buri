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
	"os"

	log "github.com/cantara/bragi/sbragi"
	"github.com/cantara/buri/download"
	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/unpack"
	"github.com/spf13/cobra"
)

var filterFlagVar = filterFlag{}

func get(afd ArtifactDownloader, unpkr Unpacker, pr download.PackageRepo, ch ConfigHandler, packageType pack.Type, artifactId, groupId string, unpack bool) {
	groupId, artifactId, artifactName, linkName, subArtifact := fixArtifactStrings(groupId, artifactId, packageType)
	repoUrl, f := ch.Config(artifactName)

	wd, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatal("while getting working dir")
	}
	dir := os.DirFS(wd)
	if unpack {
		linkName = packageType.TrimExtention(linkName)
	}
	fullPath := afd.Download(dir, pr, packageType, linkName, artifactId, groupId, repoUrl, subArtifact, f)
	if !unpack || fullPath == "" {
		return
	}
	unpkr.Unpack(dir, fullPath, linkName, packageType)
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <packageType>",
	Short: "Downloads version of software that matches filter",
	Long: `Uses the filter provided to download the specified software from Nexus.

The software will be downloaded to the working directory and unpackaged if needed.`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{string(pack.Jar), string(pack.Go), string(pack.Tar), string(pack.Zip)}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		packageType := pack.TypeFromString(args[0])
		artifactId, _ := cmd.Flags().GetString("artifact")
		groupId, _ := cmd.Flags().GetString("group")
		up, _ := cmd.Flags().GetBool("unpack")

		get(download.ArtifactDownloader{},
			unpack.Unpacker{},
			PackageRepo{},
			ViperConfigHandler{},
			packageType, artifactId, groupId,
			up,
		)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().VarP(&filterFlagVar, "filter", "f", "Filter for matching versions")
	getCmd.Flags().StringP("artifact", "a", "buri", "Artifact id of the software")
	getCmd.Flags().StringP("group", "g", "no.cantara.gotools", "Artifact group of the software")
	getCmd.Flags().BoolP("unpack", "u", false, "unpack archive")

	getCmd.RegisterFlagCompletionFunc("filter", filterCompletion)
}
