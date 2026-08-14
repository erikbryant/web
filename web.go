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

var (
	maxAttempts = 4
	retryDelay  = 500 * time.Millisecond
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

// Request makes a GET request to url, retrying transient failures.
func Request(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			if attempt == maxAttempts {
				return nil, fmt.Errorf("GET %q: %w", url, err)
			}

			time.Sleep(retryDelay)
			continue
		}

		// Return all non-5xx responses.
		if resp.StatusCode < http.StatusInternalServerError {
			return resp, nil
		}

		// This is a 5xx response.
		if attempt == maxAttempts {
			return resp, nil
		}

		// We're going to retry, so this response must be closed.
		resp.Body.Close()

		time.Sleep(retryDelay)
	}

	// The loop always returns, so we should never get here.
	return nil, fmt.Errorf("GET %q: request failed after %d attempts", url, maxAttempts)
}

// RequestBody makes a GET request and returns the response body as a string.
func RequestBody(url string, headers map[string]string) (string, error) {
	resp, err := Request(url, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body from %q: %w", url, err)
	}

	return string(body), nil
}

// RequestJSON makes a GET request and decodes the response as a JSON object.
func RequestJSON(url string, headers map[string]string) (map[string]any, error) {
	resp, err := Request(url, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var result map[string]any

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode JSON response from %q: %w", url, err)
	}

	return result, nil
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	return nil
}

// ToInt converts a value to int, or panics if the value cannot be converted.
func ToInt(val any) int {
	switch val := val.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case string:
		val = strings.ReplaceAll(val, ",", "")
		result, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot convert %q to int: %v", val, err))
		}
		return int(result)
	case float64:
		return int(val)
	default:
		panic(fmt.Sprintf("cannot convert %T (%v) to int", val, val))
	}
}

// ToInt64 converts a value to int64, or panics if the value cannot be converted.
func ToInt64(val any) int64 {
	switch val := val.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case string:
		val = strings.ReplaceAll(val, ",", "")
		result, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot convert %q to int64: %v", val, err))
		}
		return result
	case float64:
		return int64(val)
	default:
		panic(fmt.Sprintf("cannot convert %T (%v) to int64", val, val))
	}
}

// ToString converts a value to string, or panics if the value cannot be converted.
func ToString(val any) string {
	switch val := val.(type) {
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case nil:
		return ""
	default:
		panic(fmt.Sprintf("cannot convert %T (%v) to string", val, val))
	}
}

// ToFloat64 converts a value to float64, or panics if the value cannot be converted.
func ToFloat64(val any) float64 {
	switch val := val.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		val = strings.ReplaceAll(val, ",", "")
		result, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(fmt.Sprintf("cannot convert %q to float64: %v", val, err))
		}
		return result
	case float64:
		return val
	default:
		panic(fmt.Sprintf("cannot convert %T (%v) to float64", val, val))
	}
}

// MsiValue returns the value at keys in a map[string]any tree.
func MsiValue(msi any, keys []string) (any, error) {
	value := msi

	for _, key := range keys {
		object, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"cannot access key %q: value has type %T",
				key,
				value,
			)
		}

		value, ok = object[key]
		if !ok {
			return nil, fmt.Errorf("key %q not found", key)
		}
	}

	return value, nil
}

// MsiValued returns the value at keys in a map[string]any tree,
// or d if the resulting value is nil.
func MsiValued(msi any, keys []string, d any) (any, error) {
	value, err := MsiValue(msi, keys)
	if value == nil {
		return d, nil
	}
	return value, err
}
