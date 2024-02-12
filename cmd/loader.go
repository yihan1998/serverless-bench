package main

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

var (
	urlRegex = regexp.MustCompile("at URL:\nhttp://([^\n]+)")
)

func main() {
	log.Infof("Test")
}
