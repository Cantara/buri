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

	"github.com/cantara/buri/download"
	"github.com/cantara/buri/pack"
	"github.com/spf13/cobra"
)

func install(afd ArtifactDownloader, pr download.PackageRepo, ch ConfigHandler, packageType pack.Type, artifactId, groupId string) {
	groupId, artifactId, artifactName, linkName, subArtifact := fixArtifactStrings(groupId, artifactId, packageType)
	repoUrl, f := ch.Config(artifactName)

	afd.Download(os.DirFS("/usr/local/bin"), pr, packageType, linkName, artifactId, groupId, repoUrl, subArtifact, f)
}

// installCmd represents the get command
var installCmd = &cobra.Command{
	Use:   "install <packageType>",
	Short: "Installs version of software that matches filter",
	Long: `Uses the provided filter to install the specified software from Nexus.

The software will be downloaded to the working directory and unpackaged if needed.`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{string(pack.Jar), string(pack.Go)}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		packageType := pack.TypeFromString(args[0])
		artifactId, _ := cmd.Flags().GetString("artifact")
		groupId, _ := cmd.Flags().GetString("group")

		install(
			download.ArtifactDownloader{},
			PackageRepo{},
			ViperConfigHandler{},
			packageType, artifactId, groupId,
		)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().VarP(&filterFlagVar, "filter", "f", "Filter for matching versions")
	installCmd.Flags().StringP("artifact", "a", "buri", "Artifact id of the software")
	installCmd.Flags().StringP("group", "g", "no.cantara.gotools", "Artifact group of the software")

	installCmd.RegisterFlagCompletionFunc("filter", filterCompletion)
}
