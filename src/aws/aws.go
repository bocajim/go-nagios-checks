package aws

import (
	"flag"
	"fmt"
	"nagios"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/bocajim/evaler"
)

var access_id string
var secret_key string
var region string
var namespace string
var metricname string
var window int
var dimension string
var value string

func RegisterFlags() {

	flag.StringVar(&access_id, "ak", "", "AWS Access ID")
	flag.StringVar(&secret_key, "sk", "", "AWS Secret Key")
	flag.StringVar(&region, "rg", "us-east-1", "AWS Region")

	flag.StringVar(&namespace, "ns", "", "CW Namespace")
	flag.StringVar(&metricname, "mn", "", "CW MetricName")
	flag.IntVar(&window, "w", 10, "CW Time window in minutes")
	flag.StringVar(&dimension, "d", "", "CW Dimension")
	flag.StringVar(&value, "v", "", "CW Value")
}

func CheckCloudWatch() {

	if len(access_id) == 0 {
		nagios.ReturnResult(nagios.StatusUnknown, "No AWS access_id specified")
	}
	if len(secret_key) == 0 {
		nagios.ReturnResult(nagios.StatusUnknown, "No AWS secret_key specified")
	}
	if len(region) == 0 {
		nagios.ReturnResult(nagios.StatusUnknown, "No AWS region specified")
	}

	svc := cloudwatch.New(session.New(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(access_id, secret_key, "")).WithRegion(region)))

	input := &cloudwatch.GetMetricStatisticsInput{}
	input.MetricName = aws.String(metricname)
	input.Namespace = aws.String(namespace)
	input.StartTime = aws.Time(time.Now().Add(time.Minute * time.Duration(window*-1)))
	input.EndTime = aws.Time(time.Now())
	input.Dimensions = []*cloudwatch.Dimension{&cloudwatch.Dimension{Name: aws.String(dimension), Value: aws.String(value)}}
	input.Statistics = []*string{aws.String(cloudwatch.StatisticAverage)}
	input.Period = aws.Int64(int64(window * 60))
	output, err := svc.GetMetricStatistics(input)
	if err != nil {
		nagios.ReturnResult(nagios.StatusUnknown, "AWS API call error: "+err.Error())
	}
	var resultValue float64
	for _, dp := range output.Datapoints {
		if dp.Average != nil {
			resultValue = *dp.Average
		}
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
		nagios.ReturnResult(nagios.StatusCritical, "[%s] %s - %0.3f %s", value, metricname, resultValue, nagios.CriticalComparison)
	} else if isWarning {
		nagios.ReturnResult(nagios.StatusWarning, "[%s] %s - %0.3f %s", value, metricname, resultValue, nagios.WarnComparison)
	} else {
		nagios.ReturnResult(nagios.StatusOk, "[%s] %s - %0.3f", value, metricname, resultValue)
	}

}
