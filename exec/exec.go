package exec

import (
	"context"
	"errors"
	log "github.com/cantara/bragi/sbragi"
	"github.com/shirou/gopsutil/v3/process"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
)

func KillService(command []string) (killed bool) {
	commandString := strings.Join(command, " ")

	procs, err := process.Processes()
	if err != nil {
		log.WithError(err).Fatal("while getting processes")
	}
	for _, proc := range procs {
		if uids, err := proc.Uids(); err != nil || int(uids[0]) != os.Getuid() {
			continue
		}
		cmd, err := proc.Cmdline()
		if err != nil {
			log.WithError(err).Warning("while getting cmd")
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
				log.WithError(err).Error("while terminating service", "cmd", cmd)
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
		log.WithError(err).Fatal("while getting processes")
	}
	for _, proc := range procs {
		if uids, err := proc.Uids(); err != nil || int(uids[0]) != os.Getuid() {
			continue
		}
		cmd, err := proc.Cmdline()
		if err != nil {
			log.WithError(err).Warning("while getting cmd")
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

func StartService(startScript, outFile string) {
	MakeFile(outFile, "")
	err := syscall.Exec(startScript, nil, os.Environ())
	if err != nil {
		log.WithError(err).Error("while executing start script")
	}
}

func MakeFile(path, content string) {
	_, err := os.Stat(path)
	if err == nil {
		return
	}
	if !errors.Is(err, os.ErrNotExist) {
		log.WithError(err).Error("while checking if tile exists", "path", path)
		return
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.WithError(err).Error("while creating file")
		return
	}
	_, err = io.WriteString(f, content)
	if err != nil {
		log.WithError(err).Error("while writing initial content to file")
		return
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.WithError(err).Warning("while closing the newly created file")
		}
	}()
}
