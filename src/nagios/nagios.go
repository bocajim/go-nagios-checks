package nagios

import (
	"fmt"
	"os"
)

type Status string

const (
	StatusOk       = Status("OK")
	StatusWarning  = Status("WARNING")
	StatusCritical = Status("CRITICAL")
	StatusUnknown  = Status("UNKNOWN")
)

func ReturnResult(status Status, message string, values ...interface{}) {

	str := fmt.Sprintf("%s - %s\n", status, message)
	fmt.Printf(str, values...)
	switch status {
	case StatusOk:
		os.Exit(0)
	case StatusWarning:
		os.Exit(1)
	case StatusCritical:
		os.Exit(2)
	default:
		os.Exit(3)
	}
	return
}
