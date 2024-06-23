package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

var executors = []*Executor{}

func shortCommand(command string) string {
	ret := command
	parts := strings.Split(command, "\n")
	for _, i := range parts {
		i = strings.TrimLeft(i, " \t#")
		i = strings.TrimRight(i, " \t\\")
		if i != "" {
			ret = i
			break
		}
	}
	return ret
}

func runProc(cmd string, shellMethod string, dir string, v *gocui.View) error {
	writeToGuiAndUpdate(v, []byte(">> "+shortCommand(cmd)))

	ex, err := NewExecutor(shellMethod, cmd, dir)
	if err != nil {
		return err
	}
	executors = append(executors, ex)

	start := time.Now()
	state, err := ex.Run(v)
	if err != nil {
		return err
	} else if state.Error != nil {
		return state.Error
	}

	writeToGuiAndUpdate(v, []byte(fmt.Sprintf(">> Done (%s)", time.Since(start))))

	return nil
}

func stopAllProcs() {
	for _, ex := range executors {
		if err := ex.Stop(); err != nil {
			fmt.Print(err.Error())
		}
	}
}
