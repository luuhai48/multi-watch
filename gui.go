package main

import (
	"fmt"
	"math"

	"github.com/jroimartin/gocui"
)

var (
	windowMinWidth  = 60
	windowMinHeight = 20
	mainGui         *gocui.Gui
)

func writeString(v *gocui.View, line string) {
	fmt.Fprintln(v, line)
}

func writeBytes(v *gocui.View, line []byte) {
	fmt.Fprintln(v, string(line))
}

func guiUpdate() {
	if mainGui != nil {
		mainGui.Update(func(g *gocui.Gui) error {
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

		numCols := len(cmds)
		numRows := 1
		winHeight := maxY - 1
		if maxX/numCols < windowMinWidth {
			numCols = int(math.Floor(float64(maxX) / float64(windowMinWidth)))
			numRows = int(math.Ceil(float64(len(cmds)) / float64(numCols)))
			winHeight = int(math.Floor(float64(maxY) / float64(numRows)))
		}
		winWidth := int(math.Floor(float64(maxX) / float64(numCols)))

		if maxX < windowMinWidth || maxY < windowMinHeight || winWidth < windowMinWidth || winHeight < windowMinHeight {
			v, err := g.SetView("error", 0, 0, maxX-1, maxY-1)
			if err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				fmt.Fprintln(v, "Terminal window is too small")
			}
			g.SetViewOnTop("error")
			return nil
		}

		// Create windows for each command
		for i, cmd := range cmds {
			winName := fmt.Sprintf("cmd%d", i)

			if v, err := g.SetView(
				winName,
				(i%numCols)*winWidth,
				(i/numCols)*winHeight,
				(i%numCols+1)*winWidth-1,
				(i/numCols+1)*winHeight-1,
			); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}

				// window first init
				v.Wrap = true
				v.Autoscroll = true
				v.Title = shortCommand(cmd)

				// Run command and write output to view
				go func(cmd string) {
					if err := runProc(cmd, dir, v); err != nil {
						writeString(v, err.Error())
						guiUpdate()
					}
				}(cmd)
			} else {
				g.SetViewOnTop(winName)
			}
		}

		return nil
	}
}
