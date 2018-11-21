package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// returns input if it matchs with regex or empty string if not
func percentileMatch(stat string, regex *regexp.Regexp) string {
	if !regex.MatchString(stat) {
		return ""
	}
	return stat
}

// Parsing dimensions list as Shorthand syntax from parameter
func parseDimensions(dimensionsArg string) ([]*cloudwatch.Dimension, error) {
	var dimensions []*cloudwatch.Dimension
	for _, dimensionsPair := range strings.Split(dimensionsArg, " ") {
		var dimension cloudwatch.Dimension
		for _, dimensionSingle := range strings.Split(dimensionsPair, ",") {
			if strings.HasPrefix(dimensionSingle, "Name=") {
				dimension.Name = &strings.Split(dimensionSingle, "=")[1]
			} else if strings.HasPrefix(dimensionSingle, "Value=") {
				dimension.Value = &strings.Split(dimensionSingle, "=")[1]
			} else {
				return nil, errors.New("Dimensions do not match with the cloudwatch shorthand format please check https://docs.aws.amazon.com/cli/latest/reference/cloudwatch/get-metric-statistics.html")
			}
		}
		dimensions = append(dimensions, &dimension)
	}
	return dimensions, nil
}

func main() {
	const defaultDelay = "300s"

	roleArn := flag.String("role-arn", "", "AWS role ARN to assume like arn:aws:iam::myaccountid:role/myrole (optional)")
	region := flag.String("region", "", "AWS Cloudwatch region to query (mandatory)")
	namespace := flag.String("namespace", "", "AWS Cloudwatch namespace of target metric (mandatory)")
	metric := flag.String("metric", "", "AWS Cloudwatch metric name to collect (mandatory)")
	stat := flag.String("stat", "", "AWS Cloudwatch metric statistic (mandatory)")
	period := flag.Int64("period", 60, "AWS Cloudwatch metric period in seconds (optional)")
	durationString := flag.String("duration", defaultDelay, "AWS Cloudwatch metric duration as string. Ignored if \"window\" parameter is defined (optional)")
	dimensionsShorthand := flag.String("dimensions", "", "AWS Cloudwatch dimensions list to filter in Shorthand syntax as for awscli (mandatory)")
	noDataString := flag.String("no-data-value", "", "Value to return when there is no data (mandatory)")
	delayString := flag.String("delay", "300s", "AWS Cloudwatch metric delay as string. Ignored if \"window\" parameter is defined (optional)")
	window := flag.String("window", "", "AWS Cloudwatch metric window in \"duration[:delay]\" format like \"300s:300s\" (optional)")

	flag.Parse()
	if *region == "" || *namespace == "" || *metric == "" || *stat == "" || *dimensionsShorthand == "" || *noDataString == "" {
		fmt.Println("At least -namespace, -dimensions, -metric, -stat, -region and -no-data-value options must be provided")
		os.Exit(3)
	}

	// Convert nodata parameter from string (to make it possible mandatory test) to integer
	noDataValue, err := strconv.ParseInt(*noDataString, 10, 64)
	if err != nil {
		panic(err)
	}

	// Parsing dimensions list as Shorthand syntax from parameter :
	dimensions, err := parseDimensions(*dimensionsShorthand)
	if err != nil {
		fmt.Println(err)
		os.Exit(5)
	}

	if *window != "" {
		windowSlice := strings.Split(*window, ":")
		*durationString = windowSlice[0]
		if len(windowSlice) == 1 {
			*delayString = defaultDelay
		} else {
			*delayString = windowSlice[1]
		}
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigDisable,
	}))
	config := &aws.Config{
		Region: aws.String(*region),
	}

	if *roleArn != "" {
		creds := stscreds.NewCredentials(sess, *roleArn)
		config.Credentials = creds
	}

	client := cloudwatch.New(sess, config)

	duration, err := time.ParseDuration(*durationString)
	if err != nil {
		fmt.Println(err)
		os.Exit(6)
	}
	delay, err := time.ParseDuration(*delayString)
	if err != nil {
		fmt.Println(err)
		os.Exit(6)
	}
	end := time.Now().Add(-delay)
	start := end.Add(-duration)
	parameters := &cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		MetricName: aws.String(*metric),
		Namespace:  aws.String(*namespace),
		Period:     aws.Int64(*period),
		Dimensions: dimensions,
	}

	statisticsValues := []string{"SampleCount", "Average", "Sum", "Minimum", "Maximum"}
	percentileRegex := regexp.MustCompile(`p(\d{1,2}(\.\d{0,2})?|100)`)

	if stringInSlice(*stat, statisticsValues) {
		// If stat is a simple statistics
		parameters.Statistics = []*string{
			aws.String(*stat),
		}
	} else {
		switch percentileMatch(*stat, percentileRegex) {
		case *stat:
			// If stat is an extende statistics (only support percentile)
			parameters.ExtendedStatistics = []*string{
				aws.String(*stat),
			}
		default:
			fmt.Println("Unknown Cloudwatch statistics provided as stat parameter, please check https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html#Statistic")
			os.Exit(4)
		}
	}

	response, err := client.GetMetricStatistics(parameters)

	if err != nil {
		panic(err)
	}

	// Sort desc list by timestamp
	sort.Slice(response.Datapoints, func(i, j int) bool {
		return response.Datapoints[i].Timestamp.After(*response.Datapoints[j].Timestamp)
	})

	if response.Datapoints != nil {
		switch *stat {
		case "Sum":
			fmt.Println(*response.Datapoints[0].Sum)
		case "Average":
			fmt.Println(*response.Datapoints[0].Average)
		case "Maximum":
			fmt.Println(*response.Datapoints[0].Maximum)
		case "Minimum":
			fmt.Println(*response.Datapoints[0].Minimum)
		case "SampleCount":
			fmt.Println(*response.Datapoints[0].SampleCount)
		case percentileMatch(*stat, percentileRegex):
			// ExtendedStatistics
			for _, value := range response.Datapoints[0].ExtendedStatistics {
				fmt.Println(*value)
			}
		default:
			fmt.Println("Cloudwatch statistics seems to be an extendedStatistics but it seems to be an inconsistent percentile")
			os.Exit(6)

		}
	} else {
		fmt.Println(noDataValue)
	}
}
