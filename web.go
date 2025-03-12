package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Request2 makes an HTTP request of the given URL (with retry) and returns the response object
func Request2(url string, headers map[string]string) (*http.Response, error) {
	var resp *http.Response
	var err error

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("request object is nil")
	}
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	tries := 1
	for {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if tries >= 4 {
			break
		}
		time.Sleep(500 * time.Millisecond)
		tries++
	}

	return resp, err
}

// RequestBody returns the body of the HTTP response
func RequestBody(url string, headers map[string]string) (string, error) {
	resp, err := Request2(url, headers)
	if err != nil {
		return "", err
	}

	s, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(s), nil
}

// RequestJSON makes an HTTP request (with retries) of the given URL and returns the resulting JSON map
func RequestJSON(url string, headers map[string]string) (map[string]interface{}, error) {
	response, err := Request2(url, headers)
	if err != nil {
		return nil, err
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return nil, err
	}

	return jsonObject, nil
}

// ToInt converts an interface to an int if possible, otherwise panic
func ToInt(val interface{}) (result int) {
	switch val := val.(type) {
	case int:
		result = val
	case int64:
		result = int(val)
	case string:
		s := val
		s = strings.ReplaceAll(s, ",", "")
		tmp, _ := strconv.ParseInt(s, 10, 32)
		result = int(tmp)
	case float64:
		result = int(val)
	default:
		fmt.Println("Unknown type", val)
		result = val.(int) // Force a panic
	}

	return result
}

// ToInt64 converts an interface to an int if possible, otherwise panic
func ToInt64(val interface{}) (result int64) {
	switch val := val.(type) {
	case int:
		result = int64(val)
	case int64:
		result = val
	case string:
		s := val
		s = strings.ReplaceAll(s, ",", "")
		result, _ = strconv.ParseInt(s, 10, 64)
	case float64:
		result = int64(val)
	default:
		fmt.Println("Unknown type", val)
		result = val.(int64) // Force a panic
	}

	return result
}

// ToString converts an interface to a string if possible, otherwise panic
func ToString(val interface{}) (result string) {
	switch val := val.(type) {
	case int:
		result = strconv.FormatInt(int64(val), 10)
	case int64:
		result = strconv.FormatInt(val, 10)
	case string:
		result = val
	case float64:
		result = strconv.FormatFloat(val, 'f', -1, 64)
	case nil:
		result = ""
	default:
		fmt.Println("Unknown type", val)
		result = val.(string) // Force a panic.
	}

	return result
}

// ToFloat64 converts an interface to a float64 if possible, otherwise panic
func ToFloat64(val interface{}) (result float64) {
	switch val := val.(type) {
	case int:
		result = float64(val)
	case int64:
		result = float64(val)
	case string:
		s := val
		s = strings.ReplaceAll(s, ",", "")
		result, _ = strconv.ParseFloat(s, 64)
	case float64:
		result = val
	default:
		fmt.Println("Unknown type", val)
		result = val.(float64) // Force a panic.
	}

	return result
}

// MsiValue returns the value at 'keys' in a map[string]interface{} tree
func MsiValue(msi interface{}, keys []string) (interface{}, error) {
	var ok bool
	value := msi

	for _, key := range keys {
		value, ok = value.(map[string]interface{})[key]
		if !ok {
			return nil, fmt.Errorf("key '%s' not found", key)
		}
	}

	return value, nil
}

// MsiValued returns the value at 'keys' in a map[string]interface{} tree, or a default if value is nil
func MsiValued(msi interface{}, keys []string, d interface{}) (interface{}, error) {
	value, err := MsiValue(msi, keys)
	if value == nil {
		value = d
	}
	return value, err
}
