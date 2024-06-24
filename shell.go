package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/jroimartin/gocui"
)

type Executor struct {
	Shell   string
	Command string
	Dir     string

	cmd  *exec.Cmd
	stdo io.ReadCloser
	stde io.ReadCloser
	sync.Mutex
}

type ExecState struct {
	Error     error
	ErrOutput string
	ProcState string
}

func NewExecutor(shell string, command string, dir string) (*Executor, error) {
	_, err := makeCommand(shell, command, dir)
	if err != nil {
		return nil, err
	}
	return &Executor{
		Shell:   shell,
		Command: command,
		Dir:     dir,
	}, nil
}

func (e *Executor) start(v *gocui.View) (*exec.Cmd, *bytes.Buffer, *sync.WaitGroup, error) {
	e.Lock()
	defer e.Unlock()

	cmd, err := makeCommand(e.Shell, e.Command, e.Dir)
	if err != nil {
		return nil, nil, nil, err
	}
	e.cmd = cmd

	stdo, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stde, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	e.stdo = stdo
	e.stde = stde

	buff := new(bytes.Buffer)
	err = cmd.Start()
	if err != nil {
		return nil, nil, nil, err
	}
	wg := sync.WaitGroup{}
	wg.Add(2)

	go logOutput(&wg, stde, v)

	go logOutput(&wg, stdo, v)

	return cmd, buff, &wg, nil
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
	cmd, buff, wg, err := e.start(v)
	if err != nil {
		return nil, err
	}

	wg.Wait()

	ret := cmd.Wait()
	state := &ExecState{
		Error:     ret,
		ErrOutput: buff.String(),
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
	r := bufio.NewReader(fp)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			return
		}
		writeToGuiAndUpdate(v, string(line))
	}
}

func prepCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func (e *Executor) sendSignal(sig os.Signal) error {
	return syscall.Kill(-e.cmd.Process.Pid, sig.(syscall.Signal))
}

func makeCommand(shell string, command string, dir string) (*exec.Cmd, error) {
	shcmd, err := CheckShell(shell)
	if err != nil {
		return nil, err
	}
	var cmd *exec.Cmd
	switch shell {
	case "bash", "sh", "zsh":
		cmd = exec.Command(shcmd, "-c", command)
	case "powershell":
		cmd = exec.Command(shcmd, "-Command", command)
	}
	cmd.Dir = dir
	prepCmd(cmd)
	return cmd, nil
}

var ValidShells = map[string]bool{
	"bash":       true,
	"sh":         true,
	"zsh":        true,
	"powershell": true,
}

func CheckShell(shell string) (string, error) {
	if _, ok := ValidShells[shell]; !ok {
		return "", fmt.Errorf("unsupported shell: %q", shell)
	}

	switch shell {
	case "powershell":
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell", nil
		} else if _, err := exec.LookPath("pwsh"); err == nil {
			return "pwsh", nil
		} else {
			return "", fmt.Errorf("powershell/pwsh not on path")
		}
	default:
		return exec.LookPath(shell)
	}
}
