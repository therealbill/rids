// +build !syslog,!systemd

package main

import "log"

func LogInfo(msg string) error {
	log.Print(msg)
	return nil
}
