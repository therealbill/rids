package main

import (
	"os"
	"time"

	"github.com/codegangsta/cli"
	rpcclient "github.com/therealbill/redskull/rpcclient"
)

var (
	client  *rpcclient.Client
	app     *cli.App
	timeout = time.Second * 2
)

func main() {
	//first lets test being able to interact with the RS API
	app = cli.NewApp()
	app.Name = "rs-platformingester"
	app.Usage = "Interact with a Redskull cluster via the RPC API to upload instances under management to RIDS"
	app.Version = "0.5.1"
	app.EnableBashCompletion = true
	author := cli.Author{Name: "Bill Anderson", Email: "therealbill@me.com"}
	app.Authors = append(app.Authors, author)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "rpcaddr, r",
			Value:  "localhost:8001",
			Usage:  "Redskull RCP address in form 'ip:port'",
			EnvVar: "REDSKULL_RPCADDR",
		},
		cli.StringFlag{
			Name:   "serviceaddress, s",
			Value:  "localhost:443",
			Usage:  "RIDS Service address in form 'ip:port'",
			EnvVar: "RIDS_ADDRESS",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "ingest",
			Usage:  "Ingest known pods connection data",
			Action: ingestPods,
		},
	}
	app.Run(os.Args)
}

type PodMetaData struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Server   string `json:"server"`
	Plan     string `json:"plan"`
	Password string `json:"password"`
	Config   string `json:"config"`
	PodName  string `json:"podid"`
	Platform string `json:"platform"`
}

type APIResponse struct {
	Status        string            `json:"status"`
	StatusMessage string            `json:"statusmessage"`
	Data          map[string]string `json:"data"`
}
