package main

import (
	"context"
	"os"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/urfave/cli/v2"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
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
			&cli.IntFlag{
				Name:  "minwidth",
				Usage: "Minimum to display the command window. Default is 50",
				Value: 50,
			},
			&cli.IntFlag{
				Name:  "minheight",
				Usage: "Minimum to display the command window. Default is 10",
				Value: 10,
			},
		},
		Action: func(ctx *cli.Context) error {
			cmds := ctx.StringSlice("cmd")
			if len(cmds) == 0 {
				return cli.Exit("No command(s) provided.", 1)
			}

			windowMinWidth = ctx.Int("minwidth")
			windowMinHeight = ctx.Int("minheight")

			pwd, err := os.Getwd()
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			g, er := gocui.NewGui(gocui.OutputNormal)
			if er != nil {
				return cli.Exit(er.Error(), 1)
			}
			defer g.Close()

			g.Mouse = true
			mainGui = g

			g.SetManagerFunc(guiLayoutManager(cmds, pwd))

			if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, guiQuit); err != nil {
				return cli.Exit(err.Error(), 1)
			}

			if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
				return cli.Exit(err.Error(), 1)
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "exec",
				Usage: "Execute a command in the built-in shell",
				Args:  true,
				Action: func(c *cli.Context) error {
					cmd := c.Args().First()

					parser := syntax.NewParser()
					prog, err := parser.Parse(strings.NewReader(cmd), "")
					if err != nil {
						return cli.Exit(err.Error(), 1)
					}

					runner, err := interp.New(
						interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
						func(r *interp.Runner) error {
							return nil
						},
					)
					if err != nil {
						return cli.Exit(err.Error(), 1)
					}

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					runner.Reset()

					err = runner.Run(ctx, prog)
					if err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
