package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"github.com/jroimartin/gocui"
)

type Executor struct {
	Command string
	Dir     string

	cmd *exec.Cmd
	sync.Mutex
}

type ExecState struct {
	Error     error
	ProcState string
}

func NewExecutor(command string, dir string) (*Executor, error) {
	_, err := makeCommand(command, dir)
	if err != nil {
		return nil, err
	}
	return &Executor{
		Command: command,
		Dir:     dir,
	}, nil
}

func (e *Executor) start(v *gocui.View) (*exec.Cmd, *sync.WaitGroup, error) {
	e.Lock()
	defer e.Unlock()

	cmd, err := makeCommand(e.Command, e.Dir)
	if err != nil {
		return nil, nil, err
	}
	e.cmd = cmd

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go logOutput(&wg, f, v)

	return cmd, &wg, nil
}

func (e *Executor) running() bool {
	return e.cmd != nil
}

func (e *Executor) Running() bool {
	e.Lock()
	defer e.Unlock()
	return e.running()
}

func (e *Executor) reset() {
	e.Lock()
	defer e.Unlock()
	e.cmd = nil
}

func (e *Executor) Run(v *gocui.View) (*ExecState, error) {
	if e.cmd != nil {
		return nil, fmt.Errorf("already running")
	}
	cmd, wg, err := e.start(v)
	if err != nil {
		return nil, err
	}

	wg.Wait()

	ret := cmd.Wait()
	state := &ExecState{
		Error:     ret,
		ProcState: cmd.ProcessState.String(),
	}
	e.reset()
	return state, nil
}

func (e *Executor) Signal(sig os.Signal) error {
	e.Lock()
	defer e.Unlock()
	if !e.running() {
		return fmt.Errorf("executor not running")
	}
	return e.sendSignal(sig)
}

func (e *Executor) Stop() error {
	return e.Signal(os.Kill)
}

func logOutput(wg *sync.WaitGroup, fp io.ReadCloser, v *gocui.View) {
	defer wg.Done()
	for {
		r := bufio.NewReader(fp)
		for {
			line, _, err := r.ReadLine()
			if err != nil {
				return
			}
			writeBytes(v, line)
			guiUpdate()
		}
	}
}

func (e *Executor) sendSignal(sig os.Signal) error {
	return syscall.Kill(-e.cmd.Process.Pid, sig.(syscall.Signal))
}

func makeCommand(command string, dir string) (*exec.Cmd, error) {
	shcmd, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(shcmd, "exec", command)
	cmd.Dir = dir
	return cmd, nil
}
