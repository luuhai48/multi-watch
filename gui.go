package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

var (
	windowMinWidth = 50
	mainGui        *gocui.Gui
)

func writeToGuiAndUpdate(v *gocui.View, line []byte) {
	if mainGui != nil {
		mainGui.Update(func(g *gocui.Gui) error {
			fmt.Fprintln(v, string(line))
			return nil
		})
	}
}

func guiQuit(g *gocui.Gui, v *gocui.View) error {
	stopAllProcs()
	return gocui.ErrQuit
}

func guiLayoutManager(cmds []string, dir string) func(*gocui.Gui) error {
	return func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		for i, cmd := range cmds {
			// TODO: position window
			if v, err := g.SetView(
				fmt.Sprintf("cmd%d", i),
				1,
				0,
				maxX-1,
				maxY-1,
			); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				// window first init
				v.Wrap = true
				v.Autoscroll = true

				// Run command and write output to view
				go func(cmd string) {
					if err := runProc(cmd, "bash", dir, v); err != nil {
						writeToGuiAndUpdate(v, []byte(err.Error()))
					}
				}(cmd)
			}
		}

		return nil
	}
}
