package main

import (
	"os"
	"path/filepath"

	"github.com/jroimartin/gocui"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "multi-watch",
		Usage:                "Run multiple commands at once on their own window",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "cmd",
				Aliases: []string{"c"},
				Usage:   "Command to run. Can have multiple.",
			},
		},
		Action: func(ctx *cli.Context) error {
			cmds := ctx.StringSlice("cmd")
			if len(cmds) == 0 {
				return cli.Exit("No command(s) provided.", 1)
			}

			dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				cli.Exit(err.Error(), 1)
			}

			g, er := gocui.NewGui(gocui.OutputNormal)
			if er != nil {
				return cli.Exit(er.Error(), 1)
			}
			defer g.Close()
			mainGui = g

			g.SetManagerFunc(guiLayoutManager(cmds, dir))

			if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, guiQuit); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
				return cli.Exit(err.Error(), 1)
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
