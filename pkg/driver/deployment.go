package driver

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

var (
	urlRegex = regexp.MustCompile("at URL:\nhttp://([^\n]+)")
)

type Driver struct {
}

func NewDriver() {
	// return &Driver{}
	log.Infof("Test")
}

// func deployKnative() bool {
// 	cmd := exec.Command(
// 		"bash",
// 		"./pkg/driver/deploy.sh",
// 	)

// 	stdoutStderr, err := cmd.CombinedOutput()
// 	log.Debug("CMD response: ", string(stdoutStderr))

// 	if err != nil {
// 		// TODO: there should be a toggle to turn off deployment because if this is fatal then we cannot test the thing locally
// 		log.Warnf("Failed to deploy function: %v\n%s\n", err, stdoutStderr)

// 		return false
// 	}

// 	// if endpoint := urlRegex.FindStringSubmatch(string(stdoutStderr))[1]; endpoint != function.Endpoint {
// 	// 	// TODO: check when this situation happens
// 	// 	log.Debugf("Update function endpoint to %s\n", endpoint)
// 	// 	function.Endpoint = endpoint
// 	// } else {
// 	// 	function.Endpoint = fmt.Sprintf("%s.%s.%s", function.Name, namespace, bareMetalLbGateway)
// 	// 	log.Infof("Endpoint: \t%s\n", function.Endpoint)
// 	// }

// 	endpoint := urlRegex.FindStringSubmatch(string(stdoutStderr))[1]
// 	log.Infof("Endpoint: \t%s\n", endpoint)
// 	return true
// }
