package main

import "testing"

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
