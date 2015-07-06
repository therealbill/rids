package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/codegangsta/cli"
	"github.com/jmcvetta/napping"
	"github.com/therealbill/libredis/client"
	rpcclient "github.com/therealbill/redskull/rpcclient"
)

var (
	rsclient *rpcclient.Client
	app      *cli.App
	timeout  = time.Second * 2
	rc       *client.Redis
	wg       sync.WaitGroup
	ichan    chan Instance
	session  napping.Session
	urlbase  string
)

type Instance struct {
	Id     string
	Host   string
	Port   int
	Auth   string `json:"password"`
	Config string
}

func main() {
	//first lets test being able to interact with the RS API
	app = cli.NewApp()
	app.Name = "rs-dataingester"
	app.Usage = "Interact with a Redskull cluster via the RPC API to load per-instance data"
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
			Usage:  "Ingest known pods data",
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
	Extra         []Instance        `json:"extra"`
}

// getInstances() pulls instane list from the API server
func getInstances(url string) (instances []Instance) {
	//var rawbody []byte
	var resp APIResponse
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	session = napping.Session{Client: &http.Client{Transport: tr}}
	_, err := session.Get(url, nil, &resp, nil)
	if err != nil {
		log.Fatal("error on getting JSON: ", err)
		return
	}
	for _, i := range resp.Extra {
		instances = append(instances, i)
	}
	return
}

func ingestPods(c *cli.Context) {
	//var err error
	urlbase = "https://" + c.GlobalString("serviceaddress") + "/api/"
	instanceurl := "https://" + c.GlobalString("serviceaddress") + "/api/meta/platform/or"
	var err error
	rsclient, err = rpcclient.NewClient(c.GlobalString("rpcaddr"), timeout)
	if err != nil {
		log.Printf("Unable to connect to RedSkull RPC")
		return
	}
	ilist := getInstances(instanceurl)
	log.Printf("%d instances found", len(ilist))
	ichan = make(chan Instance, 64)
	for x := 0; x < 64; x++ {
		wg.Add(1)
		go sweepInstance()
	}
	for _, i := range ilist {
		ichan <- i
	}
	close(ichan)
	wg.Wait()
}
