// +build syslog,!syslog

package main

import "errors"

func LogInfo(msg string) error {
	return errors.New("system log service not available")
}
