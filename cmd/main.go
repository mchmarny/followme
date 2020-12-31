package main

import (
	"log"
	"os"

	"github.com/mchmarny/followme/internal/app"
	"github.com/mchmarny/followme/internal/worker"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var (
	// Version is the app version set at build time.
	Version string = "v0.0.1-default"
)

func main() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "key",
			Aliases:  []string{"k"},
			Usage:    "Twitter API Key",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "secret",
			Aliases:  []string{"s"},
			Usage:    "Twitter API Secret",
			Required: true,
		},
	}

	appCmd := &cli.App{
		Name:  "followme",
		Usage: "run followme",
		Commands: []*cli.Command{
			{
				Name:  "app",
				Usage: "run app",
				Flags: []cli.Flag{
					flags[0],
					flags[1],
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "app server port",
						Value:   8080,
					},
					&cli.StringFlag{
						Name:    "url",
						Aliases: []string{"u"},
						Usage:   "app server base URL",
						Value:   "http://127.0.0.1",
					},
				},
				Action: func(c *cli.Context) error {
					a, err := app.NewApp(c.String("key"),
						c.String("secret"),
						c.String("url"),
						Version,
						c.Int("port"))
					if err != nil {
						return errors.Wrap(err, "error creating new app service")
					}
					return a.Run()
				},
			},
			{
				Name:  "worker",
				Usage: "run worker",
				Flags: flags,
				Action: func(c *cli.Context) error {
					w, err := worker.NewWorker(c.String("key"),
						c.String("secret"),
						c.String("url"),
						Version)
					if err != nil {
						return errors.Wrap(err, "error creating new worker service")
					}
					return w.Run()
				},
			},
		},
	}

	err := appCmd.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
