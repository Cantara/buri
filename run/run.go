package run

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/cantara/buri/command"
	"github.com/cantara/buri/exec"

	log "github.com/cantara/bragi/sbragi"
)

func Run(dir, rawArtifactId, name, linkName, packageType string, foundNewerVersion bool) {
	hd, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Fatal("while getting home dir")
	}
	cmd := command.Create(dir, linkName, packageType)
	startScript := fmt.Sprintf("%s/scripts/start_%s.sh", hd, name)
	outFile := fmt.Sprintf("%s/%s.out", dir, name)
	argsFile := fmt.Sprintf("%s/%s.args", dir, name)
	argsFileContent := fmt.Sprintf(`#ARG's file for %s
APP_ARGS=""`, name)
	if strings.Contains(packageType, "jar") {
		argsFileContent = argsFileContent + "\nJVM_ARGS=\"\""
	}
	exec.MakeFile(argsFile, argsFileContent)
	os.Mkdir(hd+"/scripts", 0750)
	os.WriteFile(fmt.Sprintf("%s/scripts/restart_%s.sh", hd, name), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
~/scripts/kill_%[1]s.sh
sleep 5
~/scripts/start_%[1]s.sh
`, name)), 0750)
	var startScriptContent bytes.Buffer
	startScriptContent.WriteString("#!/bin/sh\n#This script is managed by BURI https://github.com/cantara/buri")
	startScriptContent.WriteString("\nsource ")
	startScriptContent.WriteString(argsFile)
	startScriptContent.WriteRune('\n')
	startScriptContent.WriteString(`echo "Extra app args: $APP_ARGS"`)
	if strings.Contains(packageType, "jar") {
		startScriptContent.WriteRune('\n')
		startScriptContent.WriteString(`echo "Extra jvm args: $JVM_ARGS"`)
	}
	startScriptContent.WriteRune('\n')
	startScriptContent.WriteString(cmd[0])
	if strings.Contains(packageType, "jar") {
		startScriptContent.WriteString(" $JVM_ARGS")
	}
	startScriptContent.WriteString(ToBashCommandString(cmd[1:]))
	startScriptContent.WriteString(" $APP_ARGS &> ")
	startScriptContent.WriteString(outFile)
	startScriptContent.WriteString(" &")
	os.WriteFile(startScript, startScriptContent.Bytes(), 0750)
	os.WriteFile(fmt.Sprintf("%s/scripts/kill_%s.sh", hd, name), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
%s -kill > /dev/null
`, strings.Join(os.Args, " "))), 0750)
	os.WriteFile(fmt.Sprintf("%s/scripts/update_%s.sh", hd, name), []byte(fmt.Sprintf(`#!/bin/sh
#This script is managed by BURI https://github.com/cantara/buri
buri run -u %s > /dev/null
`, ToBashCommandString(os.Args[3:]))), 0750)
	proc, running := exec.IsRunning(cmd[0], linkName)
	if running {
		if foundNewerVersion {
			exec.KillService(proc)
		} else {
			return
		}
	}
	exec.StartService(startScript, outFile)
}

func ToBashCommandString(cmd []string) string {
	var buf strings.Builder
	for i := range cmd {
		buf.WriteString(" \"")
		buf.WriteString(cmd[i])
		buf.WriteRune('"')
	}

	return buf.String()
}
