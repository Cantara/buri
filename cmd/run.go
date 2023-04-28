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
	"github.com/cantara/buri/run"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <packageType>",
	Short: "Executes the defined artifact",
	Long: `Executes and creates relevant scripts for the defined artifact.

If specified a new version will be downloaded.
Scripts and execution is done in relation to eXOReactions best practices.`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{string(PackageJar), string(PackageGo), string(PackageTar)}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		packageType := args[0]
		artifactIdRaw, _ := cmd.Flags().GetString("artifact")
		groupId, _ := cmd.Flags().GetString("group")

		groupId, artifactId, artifactName, linkName, subArtifact := fixArtifactStrings(groupId, artifactIdRaw, packageType)
		repoUrl, f := getConfig(artifactName)

		wd, err := os.Getwd()
		if err != nil {
			log.WithError(err).Fatal("while getting working dir")
		}

		foundNewVersion := false
		if update, _ := cmd.Flags().GetBool("update"); update {
			foundNewVersion = download.Download(wd, packageType, linkName, artifactId, groupId, repoUrl, subArtifact, f)
		}

		run.Run(wd, artifactIdRaw, artifactName, linkName, packageType, foundNewVersion)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().VarP(filterFlagVar, "filter", "f", "Filter for matching versions")
	runCmd.Flags().StringP("artifact", "a", "buri", "Artifact id of the software")
	runCmd.Flags().StringP("group", "g", "no.cantara.gotools", "Artifact group of the software")
	runCmd.Flags().BoolP("update", "u", false, "Downloads and starts the new version if there was a new version.")

	runCmd.RegisterFlagCompletionFunc("filter", filterCompletion)
}
