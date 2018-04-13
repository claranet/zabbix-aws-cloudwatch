package main

import (
	"regexp"

	"testing"
)

// test stringInSlice function which should return true if first argument exists in the slice in second argument
func TestStringInSlice(t *testing.T) {
	statisticsValues := []string{"SampleCount", "Average", "Sum", "Minimum", "Maximum"}
	cases := []struct {
		statistics string
		result     bool
	}{
		{"Minimum", true},
		{"Sum", true},
		{"Avg", false},
		{"fzepfe8_è-@)à&~", false},
		{"p99.5", false},
	}

	for _, i := range cases {
		result := stringInSlice(i.statistics, statisticsValues)
		if result != i.result {
			t.Errorf("Incorrect result for test %s, got: %t, want: %t.", i.statistics, result, i.result)
		}
	}

}

// test percentileMatch function which should return first arg test string if it match the regex pattern in second arg
func TestPercentileMatch(t *testing.T) {
	percentileRegex := regexp.MustCompile(`p(\d{1,2}(\.\d{0,2})?|100)`)
	cases := []struct {
		extendedStatistics string
		result             string
	}{
		{"p99", "p99"},
		{"p99.5", "p99.5"},
		{"99", ""},
	}

	for _, i := range cases {
		result := percentileMatch(i.extendedStatistics, percentileRegex)
		if result != i.result {
			t.Errorf("Incorrect result for test %s, got: %s, want: %s.", i.extendedStatistics, result, i.result)
		}
	}
}

// test percentileRegex which
func TestParseDimensions(t *testing.T) {
	cases := []struct {
		shorthand string
		result    bool
	}{
		{"Name=LoadBalancerName,Value=elb-test Name=AvailabilityZone,Value=eu-west-1a", true},
		{"Name=LoadBalancerName,Value=elb-test", true},
		{"Title=LoadBalancerName,Value=elb-test", false},
	}

	for _, i := range cases {
		_, err := parseDimensions(i.shorthand)
		result := true
		if err != nil {
			result = false
		}
		if result != i.result {
			t.Errorf("Incorrect result for test %s, got error: %s.", i.shorthand, err)
		}
	}
}
