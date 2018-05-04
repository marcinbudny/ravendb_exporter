package main

import "testing"

func TestParseTimeSpan(t *testing.T) {

	testCases := make(map[string]float64)
	testCases["55.11:22:33.0444000"] = 55*24*60*60 + 11*60*60 + 22*60 + 33 + 0.0444000
	testCases["55.11:22:33"] = 55*24*60*60 + 11*60*60 + 22*60 + 33
	testCases["11:22:33"] = 11*60*60 + 22*60 + 33

	for testCase, expected := range testCases {
		t.Run(testCase, func(t *testing.T) {
			actual := timeSpanToSeconds(testCase)
			if actual != expected {
				t.Errorf("Timespan string %s should parse to %fs but parsed to %fs", testCase, expected, actual)
			}
		})
	}
}
