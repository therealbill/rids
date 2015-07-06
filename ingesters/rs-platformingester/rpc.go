package main

import (
	"fmt"
	"log"
	"net/http"

	"crypto/tls"

	"github.com/codegangsta/cli"
	"github.com/jmcvetta/napping"
	rpcclient "github.com/therealbill/redskull/rpcclient"
)

func ingestPods(c *cli.Context) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client, err := rpcclient.NewClient(c.GlobalString("rpcaddr"), timeout)
	if err != nil {
		log.Fatal(err)
	}
	pods, err := client.GetPodList()
	if err != nil {
		log.Fatal(err)
	}
	var result APIResponse
	var session = napping.Session{Client: &http.Client{Transport: tr}}
	for _, podname := range pods {
		// step one, handle the master
		pod, err := client.GetPod(podname)
		if err != nil {
			log.Printf("Pod pull failed with error '%s'", err.Error())
			continue
		}
		id := fmt.Sprintf("%s:%d", pod.Master.Address, pod.Master.Port)
		url := "https://" + c.GlobalString("serviceaddress") + "/api/meta/" + id
		payload := PodMetaData{ID: id, Address: id, Password: pod.AuthToken, Config: "config", PodName: pod.Name, Platform: "or"}
		_, err = session.Put(url, &payload, &result, nil)
		if err != nil {
			log.Printf("Error on upload of %s", pod.Name)
			log.Printf("Error on PUT: %s", err.Error())
			continue
		}
		// step two, handle the slaves
		for _, si := range pod.Master.Slaves {
			payload := PodMetaData{ID: si.Name, Address: si.Name, Password: pod.AuthToken, Config: "config", PodName: pod.Name, Platform: "or"}
			url := "https://" + c.GlobalString("serviceaddress") + "/api/meta/" + si.Name
			_, err = session.Put(url, &payload, &result, nil)
			if err != nil {
				log.Printf("Error on upload of %s", pod.Name)
				log.Printf("Error on PUT: %s", err.Error())
				continue
			}
		}

	}
}
