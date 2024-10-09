// http server
// etc hosts block
// crio config
// containerd config

package main

import (
	"aws-image-proxy/lib"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	gitHash  string
	gitRef   string
	portFlag = &cli.IntFlag{
		Name:  "Port",
		Value: 432,
		Usage: "Local port to use.",
	}
	app = &cli.App{
		Name:   "AWS pull through credentials proxy",
		Usage:  "AWS pull through credentials proxy forwards requests to AWS ECR pull through caches and adds a header to allow Access based on the role of the EC2 node.",
		Action: run,
		Flags: []cli.Flag{
			portFlag,
		},
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Start the AWS pull through credentials proxy.",
				Action: run,
			},
			{
				Name:  "install",
				Usage: "Install the application on the host.",
				Subcommands: []*cli.Command{
					{
						Name:   "etc-hosts-block",
						Action: etcHosts,
					},
					{
						Name:   "containerd",
						Action: containerd,
					},
					{
						Name:   "crio",
						Action: crio,
					},
					{
						Name:   "systemd",
						Action: systemd,
					},
				},
			},
			{
				Name:   "version",
				Usage:  "Print the version.",
				Action: version,
			},
		},
	}
)

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	port := c.Int(portFlag.Name)

	log.Printf("port: %d", port)

	return lib.RunHttpServer(c.Context, port)
}

func version(c *cli.Context) error {
	_, err := fmt.Printf("version: %s\ngit commit: %s\n", gitRef, gitHash)
	if err != nil {
		return err
	}

	return nil
}

func etcHosts(c *cli.Context) error {
	return lib.EtcHostsBlock(c.Context)
}

func containerd(c *cli.Context) error {
	port := c.Int(portFlag.Name)
	return lib.InstallContainerdConfiguration(c.Context, port)
}

func crio(c *cli.Context) error {
	port := c.Int(portFlag.Name)
	return lib.InstallCrioConfiguraiton(c.Context, port)
}

func systemd(c *cli.Context) error {
	port := c.Int(portFlag.Name)
	return lib.InstallSystemdServiceConfiguraiton(port)
}
