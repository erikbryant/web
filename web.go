package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Request2 makes an HTTP request of the given URL (with retry) and returns the response object.
func Request2(url string, headers map[string]string) (resp *http.Response, err error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
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

	return
}

// Request makes an HTTP request of the given URL and returns the resulting string.
func Request(url string, headers map[string]string) (string, error) {
	resp, err := Request2(url, headers)

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(s), nil
}

// RequestJSON makes an HTTP request (with retries) of the given URL and returns the resulting JSON map.
func RequestJSON(url string, headers map[string]string) (map[string]interface{}, error) {
	response, err := Request2(url, headers)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(response.Body)
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

// ToInt translates an arbitrary type to an int if possible, otherwise panic.
func ToInt(val interface{}) (result int) {
	switch val.(type) {
	case int:
		result = val.(int)
	case int64:
		result = int(val.(int64))
	case string:
		s := val.(string)
		s = strings.ReplaceAll(s, ",", "")
		tmp, _ := strconv.ParseInt(s, 10, 32)
		result = int(tmp)
	case float64:
		result = int(val.(float64))
	default:
		fmt.Println("Unknown type", val)
		result = val.(int) // Force a panic.
	}

	return result
}

// ToInt64 translates an arbitrary type to an int if possible, otherwise panic.
func ToInt64(val interface{}) (result int64) {
	switch val.(type) {
	case int:
		result = int64(val.(int))
	case int64:
		result = val.(int64)
	case string:
		s := val.(string)
		s = strings.ReplaceAll(s, ",", "")
		tmp, _ := strconv.ParseInt(s, 10, 64)
		result = int64(tmp)
	case float64:
		result = int64(val.(float64))
	default:
		fmt.Println("Unknown type", val)
		result = val.(int64) // Force a panic.
	}

	return result
}

// ToString translates an arbitrary type to a string if possible, otherwise panic.
func ToString(val interface{}) (result string) {
	switch val.(type) {
	case int:
		result = strconv.FormatInt(int64(val.(int)), 10)
	case int64:
		result = strconv.FormatInt(val.(int64), 10)
	case string:
		result = val.(string)
	case float64:
		result = strconv.FormatFloat(val.(float64), 'f', -1, 64)
	case nil:
		result = ""
	default:
		fmt.Println("Unknown type", val)
		result = val.(string) // Force a panic.
	}

	return result
}

// ToFloat64 translates an arbitrary type to a float64 if possible, otherwise panic.
func ToFloat64(val interface{}) (result float64) {
	switch val.(type) {
	case int:
		result = float64(val.(int))
	case int64:
		result = float64(val.(int64))
	case string:
		s := val.(string)
		s = strings.ReplaceAll(s, ",", "")
		result, _ = strconv.ParseFloat(s, 64)
	case float64:
		result = val.(float64)
	default:
		fmt.Println("Unknown type", val)
		result = val.(float64) // Force a panic.
	}

	return result
}
