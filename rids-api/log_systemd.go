// +build systemd,!syslog

package main

import (
	"errors"
	"log"

	"github.com/coreos/go-systemd/journal"
)

func LogInfo(msg string) error {
	if journal.Enabled() {
		log.Print("journal enabled")
		journal.Send(msg, journal.PriInfo, nil)
		return nil
	} else {
		log.Print("journal not enabled!")
		return errors.New("systemd journal service not available")
	}
}
