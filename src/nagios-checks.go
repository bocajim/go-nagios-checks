package main

import (
	"flag"
	"graphite"
)

//common flags
var server string
var name string

func main() {

	//common flags
	flag.StringVar(&server, "s", "http://localhost", "hostname or address of the graphite server")
	flag.StringVar(&name, "n", "unknown", "friendly name given to the check")

	graphite.RegisterFlags()

	flag.Parse()

	graphite.CheckMetric(server, name)
	return
}
