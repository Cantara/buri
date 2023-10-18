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
	"github.com/cantara/buri/pack"
	"github.com/cantara/buri/runner/start"
	"github.com/cantara/buri/runner/start/command"
	"github.com/spf13/cobra"
)

// killCmd represents the kill command
var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{string(pack.Jar), string(pack.Go)}, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		packageType := pack.TypeFromString(args[0])
		artifactIdRaw, _ := cmd.Flags().GetString("artifact")
		groupId, _ := cmd.Flags().GetString("group")

		_, _, _, linkName, _ := fixArtifactStrings(groupId, artifactIdRaw, packageType)

		wd, err := os.Getwd()
		if err != nil {
			log.WithError(err).Fatal("while getting working dir")
		}

		artifactCmd := command.Create(wd, linkName, packageType)
		proc, running := start.IsRunning(artifactCmd[0], linkName)
		if !running {
			return
		}
		log.Info("process is running, killing")
		start.KillService(proc)
	},
}

func init() {
	rootCmd.AddCommand(killCmd)

	killCmd.Flags().StringP("artifact", "a", "buri", "Artifact id of the software")
	killCmd.Flags().StringP("group", "g", "no.cantara.gotools", "Artifact group of the software")
}
