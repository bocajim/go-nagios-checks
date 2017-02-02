package main

import (
	"aws"
	"flag"
	"graphite"
	"nagios"
)

//common flags
var server string
var name string
var mode string

func main() {

	flag.StringVar(&mode, "m", "", "mode (aws-cw or graphite)")

	//common flags
	flag.StringVar(&server, "s", "http://localhost", "hostname or address of the graphite server")
	flag.StringVar(&name, "n", "unknown", "friendly name given to the check")

	nagios.RegisterFlags()
	graphite.RegisterFlags()
	aws.RegisterFlags()

	flag.Parse()

	switch mode {
	case "aws-cw":
		aws.CheckCloudWatch()
	case "graphite":
		graphite.CheckMetric(server, name)
	default:
		nagios.ReturnResult(nagios.StatusCritical, "Bad mode: "+mode)
	}
	return
}
