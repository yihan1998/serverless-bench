package main

import (
	"github.com/yihan1998/serverless-bench/pkg/driver"
)

func main() {
	driver := driver.NewDriver()
	driver.deployKnative()
}
