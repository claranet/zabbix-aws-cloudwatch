package main

import (
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

func main() {

	roleArn := flag.String("role-arn", "", "AWS role ARN to assume like arn:aws:iam::myaccountid:role/myrole (optional)")
	region := flag.String("region", "", "AWS Cloudwatch region to query (mandatory)")
	namespace := flag.String("namespace", "", "AWS Cloudwatch namespace of target metric (mandatory)")
	metric := flag.String("metric", "", "AWS Cloudwatch metric name to collect (mandatory)")
	stat := flag.String("stat", "", "AWS Cloudwatch metric statistic (mandatory)")
	period := flag.Int64("period", 60, "AWS Cloudwatch metric period in seconds (optional)")
	durationString := flag.String("duration", "300s", "AWS Cloudwatch metric duration as string (optional)")
	dimensionsShorthand := flag.String("dimensions", "", "AWS Cloudwatch dimensions list to filter in Shorthand syntax as for awscli (mandatory)")
	noDataString := flag.String("no-data-value", "", "Value to return when there is no data (mandatory)")

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
	var dimensions []*cloudwatch.Dimension
	for _, dimensionsPair := range strings.Split(*dimensionsShorthand, " ") {
		var dimension cloudwatch.Dimension
		for _, dimensionSingle := range strings.Split(dimensionsPair, ",") {
			if strings.HasPrefix(dimensionSingle, "Name=") {
				dimension.Name = &strings.Split(dimensionSingle, "=")[1]
			} else if strings.HasPrefix(dimensionSingle, "Value=") {
				dimension.Value = &strings.Split(dimensionSingle, "=")[1]
			} else {
				fmt.Println("Dimensions do not match with the cloudwatch shorthand format please check https://docs.aws.amazon.com/cli/latest/reference/cloudwatch/get-metric-statistics.html")
				os.Exit(5)
			}
		}
		dimensions = append(dimensions, &dimension)
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

	now := time.Now()
	client := cloudwatch.New(sess, config)
	duration, _ := time.ParseDuration(*durationString)
	parameters := &cloudwatch.GetMetricStatisticsInput{
		StartTime:  aws.Time(now.Add(-duration)),
		EndTime:    aws.Time(now),
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
