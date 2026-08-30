package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRequest(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		wantErr        bool
		wantStatusCode int
	}{
		{
			name:           "success",
			status:         http.StatusOK,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "not found",
			status:         http.StatusNotFound,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "bad request",
			status:         http.StatusBadRequest,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
				}

				if got := r.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
					t.Errorf("X-Requested-With = %q, want %q", got, "XMLHttpRequest")
				}

				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			resp, err := Request(server.URL, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Request() error = %v, wantErr %v", err, tt.wantErr)
			}

			if resp == nil {
				t.Fatal("Request() returned nil response")
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf(
					"status code = %d, want %d",
					resp.StatusCode,
					tt.wantStatusCode,
				)
			}
		})
	}
}

func TestRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-token")
		}

		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want %q", got, "application/json")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"Accept":        "application/json",
	}

	resp, err := Request(server.URL, headers)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestInvalidURL(t *testing.T) {
	_, err := Request("://not-a-valid-url", nil)
	if err == nil {
		t.Fatal("Request() expected error, got nil")
	}
}

func TestRequestRetries(t *testing.T) {
	var attempts int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++

		if attempts < 3 {
			http.Error(w, "temporary failure", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldDelay := retryDelay
	retryDelay = 0
	defer func() {
		retryDelay = oldDelay
	}()

	resp, err := Request(server.URL, nil)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestRequestRetriesExhausted(t *testing.T) {
	var attempts int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, "server failure", http.StatusInternalServerError)
	}))
	defer server.Close()

	oldDelay := retryDelay
	retryDelay = 0
	defer func() {
		retryDelay = oldDelay
	}()

	resp, err := Request(server.URL, nil)

	if err != nil {
		t.Fatalf("Request() error = %v, want nil", err)
	}

	if resp == nil {
		t.Fatal("Request() returned nil response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf(
			"status code = %d, want %d",
			resp.StatusCode,
			http.StatusInternalServerError,
		)
	}

	if attempts != maxAttempts {
		t.Errorf("attempts = %d, want %d", attempts, maxAttempts)
	}
}

func TestRequestBody(t *testing.T) {
	const expected = `{"hello":"world"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, expected)
	}))
	defer server.Close()

	body, err := RequestBody(server.URL, nil)
	if err != nil {
		t.Fatalf("RequestBody() error = %v", err)
	}

	if body != expected {
		t.Errorf("body = %q, want %q", body, expected)
	}
}

func TestRequestBodyHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	_, err := RequestBody(server.URL, nil)
	if err == nil {
		t.Fatal("RequestBody() expected error, got nil")
	}
}

func TestRequestJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, `{
			"name": "Thunderfury",
			"level": 80,
			"preview_item": {
				"sell_price": {
					"value": 12345
				}
			}
		}`)
	}))
	defer server.Close()

	result, err := RequestJSON(server.URL, nil)
	if err != nil {
		t.Fatalf("RequestJSON() error = %v", err)
	}

	if got := result["name"]; got != "Thunderfury" {
		t.Errorf("name = %v, want %q", got, "Thunderfury")
	}

	if got := result["level"]; got != float64(80) {
		t.Errorf("level = %v, want %v", got, float64(80))
	}

	previewItem, ok := result["preview_item"].(map[string]any)
	if !ok {
		t.Fatalf("preview_item has type %T, want map[string]any", result["preview_item"])
	}

	sellPrice, ok := previewItem["sell_price"].(map[string]any)
	if !ok {
		t.Fatalf("sell_price has type %T, want map[string]any", previewItem["sell_price"])
	}

	if got := sellPrice["value"]; got != float64(12345) {
		t.Errorf("value = %v, want %v", got, float64(12345))
	}
}

func TestRequestJSONInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `this is not JSON`)
	}))
	defer server.Close()

	_, err := RequestJSON(server.URL, nil)
	if err == nil {
		t.Fatal("RequestJSON() expected error, got nil")
	}
}

func TestRequestJSONHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := RequestJSON(server.URL, nil)
	if err == nil {
		t.Fatal("RequestJSON() expected error, got nil")
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int
	}{
		{"int", int(123), 123},
		{"int64", int64(123), 123},
		{"string", "123", 123},
		{"comma string", "12,345", 12345},
		{"float64", float64(123), 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToInt(tt.in)
			if got != tt.want {
				t.Errorf("ToInt(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int64
	}{
		{"int", int(123), 123},
		{"int64", int64(123), 123},
		{"string", "123", 123},
		{"comma string", "12,345", 12345},
		{"float64", float64(123), 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToInt64(tt.in)
			if got != tt.want {
				t.Errorf("ToInt64(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"int", int(123), "123"},
		{"int64", int64(123), "123"},
		{"string", "hello", "hello"},
		{"float64", float64(123.5), "123.5"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToString(tt.in)
			if got != tt.want {
				t.Errorf("ToString(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want float64
	}{
		{"int", int(123), 123},
		{"int64", int64(123), 123},
		{"string", "123.5", 123.5},
		{"comma string", "12,345.5", 12345.5},
		{"float64", float64(123.5), 123.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToFloat64(tt.in)
			if got != tt.want {
				t.Errorf("ToFloat64(%v) = %f, want %f", tt.name, tt.in, tt.want)
			}
		})
	}
}

func TestToIntPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("ToInt() did not panic")
		}
	}()

	ToInt(struct{}{})
}

func TestToInt64Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("ToInt64() did not panic")
		}
	}()

	ToInt64(struct{}{})
}

func TestToStringPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("ToString() did not panic")
		}
	}()

	ToString(struct{}{})
}

func TestToFloat64Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("ToFloat64() did not panic")
		}
	}()

	ToFloat64(struct{}{})
}

func TestMsaValue(t *testing.T) {
	data := map[string]any{
		"preview_item": map[string]any{
			"sell_price": map[string]any{
				"value": float64(12345),
			},
		},
	}

	value, err := MsaValue(
		data,
		[]string{"preview_item", "sell_price", "value"},
	)
	if err != nil {
		t.Fatalf("MsaValue() error = %v", err)
	}

	if value != float64(12345) {
		t.Errorf("MsaValue() = %v, want %v", value, float64(12345))
	}
}

func TestMsaValueMissingKey(t *testing.T) {
	data := map[string]any{
		"preview_item": map[string]any{},
	}

	_, err := MsaValue(
		data,
		[]string{"preview_item", "sell_price", "value"},
	)
	if err == nil {
		t.Fatal("MsaValue() expected error, got nil")
	}
}

func TestMsaValueWrongType(t *testing.T) {
	data := map[string]any{
		"preview_item": "not an object",
	}

	_, err := MsaValue(
		data,
		[]string{"preview_item", "sell_price"},
	)
	if err == nil {
		t.Fatal("MsaValue() expected error, got nil")
	}
}

func TestMsaValued(t *testing.T) {
	data := map[string]any{
		"id":   json.Number("123"),
		"name": "Widget",
		"nil":  nil,
		"nested": map[string]any{
			"name": "Nested Widget",
			"nil":  nil,
		},
	}

	tests := []struct {
		name    string
		msi     any
		keys    []string
		d       any
		want    any
		wantErr string
	}{
		{
			name: "existing value",
			msi:  data,
			keys: []string{"id"},
			d:    json.Number("999"),
			want: json.Number("123"),
		},
		{
			name: "existing string",
			msi:  data,
			keys: []string{"name"},
			d:    "Default",
			want: "Widget",
		},
		{
			name: "existing nested value",
			msi:  data,
			keys: []string{"nested", "name"},
			d:    "Default",
			want: "Nested Widget",
		},
		{
			name: "existing nil uses default",
			msi:  data,
			keys: []string{"nil"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "existing nested nil uses default",
			msi:  data,
			keys: []string{"nested", "nil"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "missing key returns default",
			msi:  data,
			keys: []string{"missing"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "missing nested key returns default",
			msi:  data,
			keys: []string{"nested", "missing"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "cannot traverse string returns default",
			msi:  data,
			keys: []string{"name", "missing"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "cannot traverse nil returns default",
			msi:  data,
			keys: []string{"nil", "missing"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "nil input returns default",
			msi:  nil,
			keys: []string{"anything"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "empty path",
			msi:  data,
			keys: nil,
			d:    "Default",
			want: data,
		},
		{
			name: "empty path with nil input",
			msi:  nil,
			keys: nil,
			d:    "Default",
			want: "Default",
		},
		{
			name: "default can be nil",
			msi:  data,
			keys: []string{"missing"},
			d:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MsaValued(tt.msi, tt.keys, tt.d)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf(
					"MsaValued() = %#v, want %#v",
					got,
					tt.want,
				)
			}
		})
	}
}
