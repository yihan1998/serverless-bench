package main

import (
	"os/exec"
	"regexp"

	log "github.com/sirupsen/logrus"
)

var (
	urlRegex = regexp.MustCompile("at URL:\nhttp://([^\n]+)")
)

func deployKnative() bool {
	cmd := exec.Command(
		"bash",
		"./pkg/driver/deploy.sh",
	)

	stdoutStderr, err := cmd.CombinedOutput()
	log.Debug("CMD response: ", string(stdoutStderr))

	if err != nil {
		// TODO: there should be a toggle to turn off deployment because if this is fatal then we cannot test the thing locally
		log.Warnf("Failed to deploy function: %v\n%s\n", err, stdoutStderr)

		return false
	}

	endpoint := urlRegex.FindStringSubmatch(string(stdoutStderr))[1]
	log.Infof("Deployed function on \t%s\n", endpoint)

	return true
}

func main() {
	deployKnative()
}
