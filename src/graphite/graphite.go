package graphite

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"nagios"
	"net/http"
	"net/url"
	"strings"

	"github.com/bocajim/evaler"
)

//http://mgmt-graphite/render?target=sumSeries(stats.gauges.prod.dwopen.*-us-*.mqtt.connections.*)&from=-1hours&format=json
/*
[ {
    "target": "sumSeries(stats.gauges.prod.dwopen.*-us-*.mqtt.connections.*)",
    "datapoints": [[10607.0, 1471190160], [10607.0, 1471190220], [10607.0, 1471190280], [10608.0, 1471190340], [10611.0, 1471190400], [10609.0, 1471190460], [10609.0, 1471190520], [10611.0, 1471190580], [10608.0, 1471190640], [10605.0, 1471190700], [10609.0, 1471190760], [10608.0, 1471190820], [10607.0, 1471190880], [10606.0, 1471190940], [10609.0, 1471191000], [10605.0, 1471191060], [10605.0, 1471191120], [10605.0, 1471191180], [10607.0, 1471191240], [10609.0, 1471191300], [10608.0, 1471191360], [10609.0, 1471191420], [10606.0, 1471191480], [10607.0, 1471191540], [10609.0, 1471191600], [10606.0, 1471191660], [10609.0, 1471191720], [10611.0, 1471191780], [10609.0, 1471191840], [10610.0, 1471191900], [10610.0, 1471191960], [10611.0, 1471192020], [10607.0, 1471192080], [10610.0, 1471192140], [10610.0, 1471192200], [10615.0, 1471192260], [10614.0, 1471192320], [10616.0, 1471192380], [10617.0, 1471192440], [10614.0, 1471192500], [10611.0, 1471192560], [10613.0, 1471192620], [10615.0, 1471192680], [10615.0, 1471192740], [10617.0, 1471192800], [10619.0, 1471192860], [10616.0, 1471192920], [10619.0, 1471192980], [10620.0, 1471193040], [10619.0, 1471193100], [10617.0, 1471193160], [10621.0, 1471193220], [10623.0, 1471193280], [10621.0, 1471193340], [10624.0, 1471193400], [10623.0, 1471193460], [10619.0, 1471193520], [10623.0, 1471193580], [10621.0, 1471193640], [10619.0, 1471193700]]
  }
]
*/

var metric string
var period string
var scale string
var aggregate string
var ignoreUnknown bool

type Result struct {
	Target     string      `json:"target"`
	Datapoints [][]float64 `json:"datapoints"`
}

func RegisterFlags() {
	flag.StringVar(&metric, "gm", "", "stastic to measure")
	flag.StringVar(&period, "gp", "-1hours", "time period to measure")
	flag.StringVar(&scale, "gs", "1", "scale value before comparing")
	flag.StringVar(&aggregate, "ga", "avg", "aggregation (avg, min, max)")
	flag.BoolVar(&ignoreUnknown, "gu", false, "ignore unknown")
}

func CheckMetric(server, name string) {

	if name == "unknown" {
		name = ""
	} else {
		name = name + " - "
	}

	if len(server) == 0 {
		nagios.ReturnResult(nagios.StatusUnknown, "Bad server: "+server)
	}
	if len(metric) == 0 {
		nagios.ReturnResult(nagios.StatusUnknown, "No metric specified")
	}

	url := server + "/render?target=" + url.QueryEscape(metric) + "&from=" + period + "&format=json"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		nagios.ReturnResult(nagios.StatusUnknown, err.Error())
	}
	client := http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		nagios.ReturnResult(nagios.StatusUnknown, err.Error())
	}
	if rsp.StatusCode != 200 {
		nagios.ReturnResult(nagios.StatusUnknown, "Invalid HTTP status code: %s (%d)", rsp.Status, rsp.StatusCode)
	}
	body, _ := ioutil.ReadAll(rsp.Body)

	var objs []Result
	err = json.Unmarshal(body, &objs)
	if err != nil {
		nagios.ReturnResult(nagios.StatusUnknown, "Could not parse graphite response: ["+string(body)+"] "+err.Error())
	}

	if len(objs) == 0 {
		if ignoreUnknown {
			nagios.ReturnResult(nagios.StatusOk, "No data returned for query.")
		} else {
			nagios.ReturnResult(nagios.StatusUnknown, "No data returned for query.")
		}
	}

	if len(objs) > 1 {
		objs = objs[:len(objs)-1]
	}

	obj := objs[0]

	resultValue := 0.0

	switch aggregate {
	case "sum":
		valSum := 0.0
		for _, val := range obj.Datapoints {
			valSum += val[0]
		}
		resultValue = valSum
	default: //avg
		valSum := 0.0
		for _, val := range obj.Datapoints {
			valSum += val[0]
		}
		resultValue = valSum / float64(len(obj.Datapoints))
	}

	if scale != "1" {

		if idx := strings.LastIndex(scale, "/"); idx != -1 {
			scale = scale[idx:]
		}
		resultScaled, err := evaler.Eval(fmt.Sprintf("%f%s", resultValue, scale))
		if err != nil {
			nagios.ReturnResult(nagios.StatusUnknown, "Could not evaluate scale: "+err.Error())
		}
		resultValue = evaler.BigratToFloat(resultScaled)
	}

	resultExprWarn, err := evaler.Eval(fmt.Sprintf("%f%s", resultValue, nagios.WarnComparison))
	if err != nil {
		nagios.ReturnResult(nagios.StatusUnknown, "Could not evaluate warning: "+err.Error())
	}
	isWarning := true
	if evaler.BigratToFloat(resultExprWarn) == 0.0 {
		isWarning = false
	}

	resultExprCritical, err := evaler.Eval(fmt.Sprintf("%f%s", resultValue, nagios.CriticalComparison))
	if err != nil {
		nagios.ReturnResult(nagios.StatusUnknown, "Could not evaluate critical: "+err.Error())
	}
	isCritical := true
	if evaler.BigratToFloat(resultExprCritical) == 0.0 {
		isCritical = false
	}

	if isCritical {
		nagios.ReturnResult(nagios.StatusCritical, "%s%0.3f %s", name, resultValue, nagios.CriticalComparison)
	} else if isWarning {
		nagios.ReturnResult(nagios.StatusWarning, "%s%0.3f %s", name, resultValue, nagios.WarnComparison)
	} else {
		nagios.ReturnResult(nagios.StatusOk, "%s%0.3f", name, resultValue)
	}

	return
}
