package web

import (
	"testing"
)

// TODO: create local test HTTP server to return text and JSON responses.

func TestToInt(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected int
	}{
		{int(9), 9},
		{int64(121), 121},
		{string("23"), 23},
		{string("2,300"), 2300},
		{string("2,300,999"), 2300999},
		{float64(99.4), 99},
	}

	for _, testCase := range testCases {
		answer := ToInt(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}

func TestToInt64(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected int64
	}{
		{int(9), 9},
		{int64(121121121121), 121121121121},
		{string("2323232323"), 2323232323},
		{string("2,323,232,323"), 2323232323},
		{float64(9988776655.4), 9988776655},
	}

	for _, testCase := range testCases {
		answer := ToInt64(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}

func TestToString(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected string
	}{
		{int(9), "9"},
		{int64(121), "121"},
		{string("23"), "23"},
		{float64(99.4), "99.4"},
		{nil, ""},
	}

	for _, testCase := range testCases {
		answer := ToString(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}

func TestToFloat64(t *testing.T) {
	testCases := []struct {
		value    interface{}
		expected float64
	}{
		{int(9), 9},
		{int64(121), 121},
		{string("23.45"), 23.45},
		{string("2300.123"), 2300.123},
		{string("2,300.123"), 2300.123},
		{string("2,000,300.123"), 2000300.123},
		{float64(99.4), 99.4},
	}

	for _, testCase := range testCases {
		answer := ToFloat64(testCase.value)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.value, testCase.expected, answer)
		}
	}
}

func TestMsiValue(t *testing.T) {
	testCases := []struct {
		msi         interface{}
		keys        []string
		expected    interface{}
		shouldError bool
	}{
		{nil, nil, nil, false},
		{map[string]interface{}{}, []string{"bad_key"}, nil, true},
		{map[string]interface{}{
			"good_key": nil,
		}, []string{"good_key"}, nil, false},
		{map[string]interface{}{
			"good_key": 5,
		}, []string{"good_key"}, 5, false},
		{map[string]interface{}{
			"good_key": "five",
		}, []string{"good_key"}, "five", false},
	}
	for _, testCase := range testCases {
		answer, err := MsiValue(testCase.msi, testCase.keys)
		if err == nil && testCase.shouldError {
			t.Errorf("ERROR: For %v should have errored", testCase.keys)
		}
		if err != nil && !testCase.shouldError {
			t.Errorf("ERROR: For %v got unexpected error %v", testCase.keys, err)
		}
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.keys, testCase.expected, answer)
		}
	}
}

func TestMsiValued(t *testing.T) {
	testCases := []struct {
		msi         interface{}
		keys        []string
		def         interface{}
		expected    interface{}
		shouldError bool
	}{
		{nil, nil, nil, nil, false},
		{nil, nil, "notFound", "notFound", false},
		{map[string]interface{}{}, []string{"bad_key"}, nil, nil, true},
		{map[string]interface{}{
			"good_key": nil,
		}, []string{"good_key"}, "gotNil", "gotNil", false},
		{map[string]interface{}{
			"good_key": 5,
		}, []string{"good_key"}, 99, 5, false},
		{map[string]interface{}{
			"good_key": "five",
		}, []string{"good_key"}, "four", "five", false},
	}
	for _, testCase := range testCases {
		answer, err := MsiValued(testCase.msi, testCase.keys, testCase.def)
		if err == nil && testCase.shouldError {
			t.Errorf("ERROR: For %v should have errored", testCase.keys)
		}
		if err != nil && !testCase.shouldError {
			t.Errorf("ERROR: For %v got unexpected error %v", testCase.keys, err)
		}
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.keys, testCase.expected, answer)
		}
	}
}
