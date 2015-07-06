package main // import "github.com/therealbill/rids/rids-api"

import (
	"encoding/base64"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/bitly/go-simplejson"
	"github.com/codegangsta/cli"
	"github.com/therealbill/libredis/client"
	"github.com/therealbill/rids/structures"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

var (
	rc  *client.Redis
	app *cli.App
)

var counters = expvar.NewMap("counters")

type RTGData struct {
	name          string
	connstring    string
	auth          string
	configcommand string
}

type APIResponse struct {
	Status        string                 `json:"status"`
	StatusMessage string                 `json:"statusmessage"`
	Data          map[string]interface{} `json:"data"`
	Extra         interface{}            `json:"extra"`
}

type ConnData map[string]interface{}

func getConnectionInfoForInstance(id string) map[string]interface{} {
	config, _ := rc.HGet(id, "config")
	address, _ := rc.HGet(id, "address")
	password, _ := rc.HGet(id, "password")
	depass, _ := decodePassword(string(password))
	var data ConnData
	data = make(map[string]interface{})
	data["config"] = string(config)
	data["password"] = string(depass)
	data["address"] = string(address)
	data["id"] = id
	return data
}

func GetInstancesByPlatform(c web.C, w http.ResponseWriter, r *http.Request) {
	platform := "instances:" + c.URLParams["id"]
	w.Header().Set("Content-Type", "application/vnd.api+json")
	var response APIResponse
	listing, err := rc.SMembers(platform)
	if err != nil {
		log.Printf("Unable to connect to storage, err: %s", err.Error())
	}
	var il []ConnData
	for _, id := range listing {
		il = append(il, getConnectionInfoForInstance(id))
	}
	data := make(map[string]interface{})
	data["instancecount"] = fmt.Sprintf("%d", len(il))
	data["instances"] = il
	response.Data = data
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("[GetInstancesByPlatform] Error encoding response object: %s", err.Error())
	}
}

func GetIndices(c web.C, w http.ResponseWriter, r *http.Request) {
	var response APIResponse
	listing, err := rc.SMembers("indices")
	if err != nil {
		log.Printf("Unable to connect to storage, err: %s", err.Error())
	}
	results := make(map[string]interface{})
	results["Indices"] = listing
	response.Data = results
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("[GetInstancesByIndex] Error encoding response object: %s", err.Error())
	}
}
func GetPlatforms(c web.C, w http.ResponseWriter, r *http.Request) {
	var response APIResponse
	listing, err := rc.SMembers("platforms")
	if err != nil {
		log.Printf("Unable to connect to storage, err: %s", err.Error())
	}
	results := make(map[string]interface{})
	results["platforms"] = listing
	response.Data = results
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("[GetInstancesByIndex] Error encoding response object: %s", err.Error())
	}
}

func GetInstancesByIndex(c web.C, w http.ResponseWriter, r *http.Request) {
	index := "index:" + c.URLParams["k"] + ":" + c.URLParams["v"]
	var response APIResponse
	listing, err := rc.SMembers(index)
	if err != nil {
		log.Printf("Unable to connect to storage, err: %s", err.Error())
	}
	results := make([]ConnData, len(listing))
	for x, id := range listing {
		//response.Extra.([]ConnData)[x] = getConnectionInfoForInstance(id)
		results[x] = getConnectionInfoForInstance(id)
	}
	response.Data = make(map[string]interface{})
	response.Data["Instances"] = results
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("[GetInstancesByIndex] Error encoding response object: %s", err.Error())
	}
}

func GetInstanceConnectionInfo(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	id := c.URLParams["id"]
	var response APIResponse
	response.Data = getConnectionInfoForInstance(id)
	response.Status = "success"
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("[GetInstanceConnectionInfo] Error encoding response object: %s", err.Error())
	}
}

func RemoveInstance(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Print("RemoveInstance called")
	w.Header().Set("Content-Type", "application/vnd.api+json")
	id := c.URLParams["id"]
	var reply APIResponse
	platform, err := rc.HGet(id, "platform")
	if err != nil {
		reply.Status = "ERROR"
		reply.StatusMessage = err.Error()
		reply.Data = make(map[string]interface{})
		reply.Data["instance_id"] = id
		reply.Data["platform"] = "unknown"
		log.Printf("Error on RemoveInstance/%s: %s", id, err.Error())
	}
	pkey := fmt.Sprintf("instances:%s", platform)
	rc.SRem("instances", id)
	rc.SRem(pkey, id)
	err = json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Printf("[RemoveInstance] Error encoding reply object: %s", err.Error())
	}
}

func StoreData(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	counters.Add("app.calls.StoreData", 1)
	id := c.URLParams["id"]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Body: '%s'", body)
		log.Printf("Error reading body: '%s'", err.Error())
		http.Error(w, "unable to read content from request", http.StatusNotAcceptable)
		return
	}

	js, err := simplejson.NewJson(body)
	if err != nil {
		log.Printf("Error reading json: '%s'", err.Error())
		http.Error(w, "unable to store data", http.StatusNotAcceptable)
		return
	}
	data, err := js.Map()
	if err != nil {
		log.Printf("Error reading json: '%s'", err.Error())
		http.Error(w, "unable to store data", http.StatusNotAcceptable)
		return
	}
	log.Printf("data in storedata: %+v", data)
	for k, vi := range data {
		v := vi.(string)
		doIndex, _ := rc.SIsMember("indices", k)
		if doIndex {
			rc.SAdd("index:"+k+":"+v, id)
		}
		if k == "port" {
			p, err := js.Get("port").Int()
			if err != nil {
				log.Printf("Error reading json: '%s'", err.Error())
				http.Error(w, "unable to store data", http.StatusNotAcceptable)
				return
			}
			rc.HSet(id, k, fmt.Sprintf("%d", p))
		} else if k == "password" {
			pw, err := encodePassword(v)
			if err != nil {
				log.Printf("Unable to encode password: %s", err.Error())
			} else {
				rc.HSet(id, k, string(pw))
			}
		} else {
			rc.HSet(id, k, v)
		}
		counters.Add("redis.ops.hset", 1)
	}
	rc.SAdd("instances:all", id)
	counters.Add("redis.ops.sadd", 1)
	p, platid_exists := data["platform"].(string)
	if platid_exists {
		counters.Add("redis.ops.sadd", 1)
		rc.SAdd("instances:"+p, id)
		rc.SAdd("platforms", "custom")
		rc.HSet(id, "platform", "custom")
	} else {
		if data["config"] != "config" {
			counters.Add("redis.ops.sadd", 1)
			rc.SAdd("instances:custom", id)
			rc.SAdd("platforms", "custom")
			rc.HSet(id, "platform", "custom")
		} else {
			counters.Add("redis.ops.sadd", 1)
			rc.SAdd("instances:unknown", id)
			rc.SAdd("platforms", "unnknown")
			rc.HSet(id, "platform", "unknown")
		}
	}
	return
}

func StoreMetrics(c web.C, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	counters.Add("app.calls.StoreMetrics", 1)
	id := c.URLParams["id"]
	body, err := ioutil.ReadAll(r.Body)
	var metricsrequest structures.InstanceDataJSON
	err = json.Unmarshal(body, &metricsrequest)
	if err != nil {
		log.Printf("Error reading json: '%s'", err.Error())
		http.Error(w, "unable to store data", http.StatusNotAcceptable)
		return
	}
	_ = id
	_, err = rc.HSet(id, "info_all", metricsrequest.Info)
	var ckey string
	for k, v := range metricsrequest.Config {
		if k == "port" {
			ckey = "iport"
		} else {
			ckey = k
		}
		// skip requirepass
		if k == "requirepass" {
			continue
		}
		rc.HSet(id, ckey, v)
	}
	return
}

func init() {
}

func main() {

	app = cli.NewApp()
	app.Name = "rids"
	app.Usage = "A daemon for storing redis instance data"
	app.Version = "v0.1.0"
	app.EnableBashCompletion = true
	author := cli.Author{Name: "Bill Anderson", Email: "therealbill@me.com"}
	app.Authors = append(app.Authors, author)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "redishost, rh",
			Value:  "localhost",
			Usage:  "Hostname/IP of the Redis storage service",
			EnvVar: "REDISHOST",
		},
		cli.IntFlag{
			Name:   "redisport, rp",
			Value:  6379,
			Usage:  "Port for the Redis storage service",
			EnvVar: "REDISPORT",
		},
		cli.StringFlag{
			Name:   "redisauth, ra",
			Value:  "",
			Usage:  "Authentication token for the Redis storage service",
			EnvVar: "REDISAUTH",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "serve",
			Usage:  "run the service",
			Action: Serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "certfile, f",
					Value:  "/etc/rids/cert.pem",
					Usage:  "Location of the certificate PEM file",
					EnvVar: "CERTFILE",
				},
				cli.StringFlag{
					Name:   "keyfile, k",
					Value:  "/etc/rids/key.pem",
					Usage:  "Location of the key PEM file",
					EnvVar: "KEYFILE",
				},
				cli.StringFlag{
					Name:   "listenon, l",
					Value:  "0.0.0.0",
					Usage:  "IP to listen on",
					EnvVar: "LISTENON",
				},
				cli.IntFlag{
					Name:   "port, p",
					Value:  443,
					Usage:  "Port to listen on",
					EnvVar: "PORT",
				},
				cli.BoolFlag{
					Name:   "plain",
					Usage:  "Do NOT over TLS",
					EnvVar: "NOTLS",
				},
				cli.StringFlag{
					Name:   "metip, m",
					Value:  "0.0.0.0",
					Usage:  "IP to serve metrics on",
					EnvVar: "METRICSIP",
				},
				cli.IntFlag{
					Name:   "metport",
					Value:  1443,
					Usage:  "Port to serve metrics on",
					EnvVar: "METRICSPORT",
				},
			},
		},
	}
	app.Run(os.Args)
}

func Serve(c *cli.Context) {
	log.Print("RemoveInstance called")
	var err error
	LogInfo(fmt.Sprintf("Connect to %s:%d", c.GlobalString("redishost"), c.GlobalInt("redisport")))
	rc, err = client.Dial(c.GlobalString("redishost"), c.GlobalInt("redisport"))
	if err != nil {
		log.Fatal("Unable to connect to store")
	}
	if c.GlobalString("redisauth") > "" {
		res, err := rc.ExecuteCommand("auth", c.GlobalString("redisauth"))
		if err != nil {
			log.Fatal("Unable to authenticate to store")
		}
		log.Printf("%+v", res)
	}

	flag.Set("bind", fmt.Sprintf(":%d", c.Int("port")))
	r := web.New()
	r.Get("/api/meta/:id", GetInstanceConnectionInfo)
	r.Put("/api/meta/:id", StoreData)
	r.Get("/api/meta/platform/:id", GetInstancesByPlatform)
	r.Get("/api/meta/platform/", GetPlatforms)
	r.Get("/api/meta/index/:k/:v", GetInstancesByIndex)
	r.Get("/api/meta/index/", GetIndices)
	r.Delete("/api/meta/:id", RemoveInstance)
	r.Put("/api/info/:id", StoreMetrics)
	log.Printf("NOTLS: %t", c.Bool("plain"))
	listenaddr := fmt.Sprintf("%s:%d", c.String("listenon"), c.Int("port"))
	log.Printf("Listening at %s", listenaddr)
	if c.Bool("plain") {
		go graceful.ListenAndServe(listenaddr, r)
		LogInfo("Serving on HTTP")
	} else {
		LogInfo(fmt.Sprintf("Serving on HTTPS @ %s", listenaddr))
		go graceful.ListenAndServeTLS(listenaddr, c.String("certfile"), c.String("keyfile"), r)
	}
	if err != nil {
		log.Printf("Unable to start. Error: %s", err.Error())
	}

	// fire up metrics server
	metrics_address := fmt.Sprintf("%s:%d", c.String("metip"), c.Int("metport"))
	log.Print("Metrics available at: ", metrics_address)
	if err := http.ListenAndServe(metrics_address, nil); err != nil {
		log.Print("Startup error:", err.Error())
	}
	log.Print("Exiting...")
}

func base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}

func base64Decode(src []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(src))
}

func encodePassword(pass string) ([]byte, error) {
	encoded := base64Encode([]byte(pass))
	return encoded, nil
}
func decodePassword(pass string) ([]byte, error) {
	decoded, err := base64Decode([]byte(pass))
	return decoded, err
}
