package exec

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"

	log "github.com/cantara/bragi/sbragi"
)

func KillService(proc *process.Process) (killed bool) {
	cmd, err := proc.Cmdline()
	if err != nil {
		log.WithError(err).Warning("while getting cmd")
	}
	log.Info("killing", cmd)

	ctxTerm, cancelTerm := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelTerm()
	err = proc.TerminateWithContext(ctxTerm)
	if err != nil {
		log.WithError(err).Warning("while trying to terminate process", "cmd", cmd)
	} else {
		killed = waitKilled(proc, cmd, ctxTerm)
		if killed {
			return
		}
	}

	ctxKill, cancelKill := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelKill()
	err = proc.KillWithContext(ctxKill)
	if err != nil {
		log.WithError(err).Error("while killing process", "cmd", cmd)
	}
	killed = waitKilled(proc, cmd, ctxKill)
	if killed {
		return
	}
	return
}

func waitKilled(proc *process.Process, cmd string, ctx context.Context) (killed bool) {
	t := time.NewTicker(50 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-t.C:
			running, err := proc.IsRunning()
			if err != nil {
				log.WithError(err).Error("while checking if process is still running", "cmd", cmd)
				return true
			}
			if !running {
				return true
			}
		}
	}
}

func IsRunning(base, linkName string) (proc *process.Process, running bool) {
	procs, err := process.Processes()
	if err != nil {
		log.WithError(err).Fatal("while getting processes")
	}
	for _, proc = range procs {
		if uids, err := proc.Uids(); err != nil || !isOwner(uids) {
			if !isOwner(uids) {
				log.Trace("skipping because owner was not this service user", "process", proc)
			}
			continue
		}
		log.Trace("cheking because owner was this service user", "process", proc)
		cmd, err := proc.Cmdline()
		if err != nil {
			log.WithError(err).Warning("while getting cmd")
			continue
		}
		if !strings.HasPrefix(cmd, base) {
			continue
		}
		if !strings.Contains(cmd, linkName) {
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

func isOwner(ids []int32) bool {
	owner := int32(os.Geteuid())
	for _, i := range ids {
		if i == owner {
			return true
		}
	}
	return false
}
