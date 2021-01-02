package main

import (
	"log"
	"os"

	"github.com/mchmarny/followme/internal/app"
	"github.com/mchmarny/followme/internal/data"
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
			EnvVars:  []string{"TWITTER_CONSUMER_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "secret",
			Aliases:  []string{"s"},
			Usage:    "Twitter API Secret",
			EnvVars:  []string{"TWITTER_CONSUMER_SECRET"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Data file path",
			EnvVars: []string{"DATA_FILE_PATH"},
			Value:   data.GetDefaultDBFilePath(),
		},
	}

	appCmd := &cli.App{
		Name:        "followme",
		Description: "Chart and monitor Twitter followers and unfollowers across multiple accounts.",
		Authors: []*cli.Author{
			{
				Name:  "Mark Chmarny",
				Email: "followme@chmarny.com",
			},
		},
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:  "app",
				Usage: "run app",
				Flags: []cli.Flag{
					flags[0],
					flags[1],
					flags[2],
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "app server port",
						EnvVars: []string{"APP_PORT"},
						Value:   8080,
					},
					&cli.StringFlag{
						Name:    "url",
						Aliases: []string{"u"},
						Usage:   "app server base URL",
						EnvVars: []string{"APP_URL"},
						Value:   "http://127.0.0.1",
					},
					&cli.BoolFlag{
						Name:    "dev",
						Aliases: []string{"d"},
						Usage:   "Developer mode",
						EnvVars: []string{"DEV_MODE"},
						Value:   false,
					},
				},
				Action: func(c *cli.Context) error {
					a, err := app.NewApp(c.String("file"), c.String("key"),
						c.String("secret"), c.String("url"), Version, c.Int("port"), c.Bool("dev"))
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
					w, err := worker.NewWorker(c.String("file"), c.String("key"),
						c.String("secret"), c.String("url"), Version)
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
