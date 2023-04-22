package exec

import (
	"context"
	"fmt"
	log "github.com/cantara/bragi"
	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"os/exec"
	"strings"
	"time"
)

func KillService(command []string) (killed bool) {
	commandString := strings.Join(command, " ")

	procs, err := process.Processes()
	if err != nil {
		log.AddError(err).Fatal("while getting processes")
	}
	for _, proc := range procs {
		if uids, err := proc.Uids(); err != nil || int(uids[0]) != os.Getuid() {
			continue
		}
		cmd, err := proc.Cmdline()
		if err != nil {
			log.AddError(err).Warning("while getting cmd")
			continue
		}
		if cmd != commandString {
			continue
		}
		killed = true
		log.Info("killing", cmd)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		err = proc.TerminateWithContext(ctx)
		cancel() //FIXME: For some reason this is not waiting for a kill
		if err != nil {
			err = proc.Kill()
			if err != nil {
				log.AddError(err).Error("while terminating service", "cmd", cmd)
			}
			//TODO: Wait for killed
		}
		break
	}
	return
}

func IsRunning(command []string) (running bool) {
	commandString := strings.Join(command, " ")

	procs, err := process.Processes()
	if err != nil {
		log.AddError(err).Fatal("while getting processes")
	}
	for _, proc := range procs {
		if uids, err := proc.Uids(); err != nil || int(uids[0]) != os.Getuid() {
			continue
		}
		cmd, err := proc.Cmdline()
		if err != nil {
			log.AddError(err).Warning("while getting cmd")
			continue
		}
		if cmd != commandString {
			continue
		}
		running = true
		break
	}
	return
}

func StartService(command []string, artifactId, linkName, wd string) {
	stdOut, err := os.OpenFile(fmt.Sprintf("%s/%sOut", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	stdErr, err := os.OpenFile(fmt.Sprintf("%s/%sErr", wd, artifactId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	var cmd *exec.Cmd
	if len(command) == 1 {
		cmd = exec.Command(command[0])
	} else {
		cmd = exec.Command(command[0], command[1:]...)
	}
	var envMap map[string]string
	envMap, err = godotenv.Read(".env." + strings.TrimSuffix(linkName, ".jar")) //, strings.TrimSuffix(linkName, ".jar")+".env")
	if err != nil {
		log.AddError(err).Info("while reading env files")
	}
	env := make([]string, len(envMap))
	i := 0
	for k, v := range envMap {
		env[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	cmd.Env = append(cmd.Environ(), env...)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	log.Debug(cmd)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
