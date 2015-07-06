package main

import (
	"log"

	"github.com/therealbill/libredis/client"
	"github.com/therealbill/ridss/structures"
)

//func sweepInstance(i Instance) error {
func sweepInstance() {
	var tc *client.Redis
	var err error
	defer wg.Done()
	var result APIResponse
	for i := range ichan {
		var payload structures.InstanceDataJSON
		url := urlbase + "info/" + i.Id
		tc, err = client.DialWithConfig(&client.DialConfig{Address: i.Id, Password: i.Auth})
		if err != nil {
			payload.Error = err.Error()
			_, err = session.Put(url, &payload, &result, nil)
			if err != nil {
				log.Printf("Error on pull_error API upload: %s", err.Error())
			}
			continue
		}
		inf, _ := tc.InfoString("all")
		payload.Info = inf
		if i.Config > "" {
			cfgres, err := tc.ExecuteCommand(i.Config, "GET", "*")
			if err != nil {
				log.Printf("Error on config pull: %s", err.Error())
				continue
			}
			chash, err := cfgres.HashValue()
			payload.Config = chash
		} else {
			log.Printf("Instance %s lacks a config entry", i.Id)
		}
		_, err = session.Put(url, &payload, &result, nil)
		if err != nil {
			log.Printf("Error on metrics API upload: %s", err.Error())
		}
	}
}
