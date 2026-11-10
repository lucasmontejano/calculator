package httpapi

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOperations(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		status     int
		wantResult *float64 // expected "result" field, when status is 200
		wantError  string   // expected substring of the "error" field, when there is an error
	}{
		// Binary operations
		{name: "add", method: http.MethodPost, path: "/api/operations/add", body: `{"a":2,"b":3}`, status: http.StatusOK, wantResult: floatPtr(5)},
		{name: "subtract", method: http.MethodPost, path: "/api/operations/subtract", body: `{"a":10,"b":4}`, status: http.StatusOK, wantResult: floatPtr(6)},
		{name: "multiply", method: http.MethodPost, path: "/api/operations/multiply", body: `{"a":-3,"b":4}`, status: http.StatusOK, wantResult: floatPtr(-12)},
		{name: "divide", method: http.MethodPost, path: "/api/operations/divide", body: `{"a":1,"b":4}`, status: http.StatusOK, wantResult: floatPtr(0.25)},
		{name: "divide by zero", method: http.MethodPost, path: "/api/operations/divide", body: `{"a":1,"b":0}`, status: http.StatusBadRequest, wantError: "division by zero"},
		{name: "power", method: http.MethodPost, path: "/api/operations/power", body: `{"a":2,"b":10}`, status: http.StatusOK, wantResult: floatPtr(1024)},
		{name: "power with non-real result", method: http.MethodPost, path: "/api/operations/power", body: `{"a":-8,"b":0.5}`, status: http.StatusUnprocessableEntity, wantError: "not a real number"},
		{name: "power with infinite result", method: http.MethodPost, path: "/api/operations/power", body: `{"a":0,"b":-1}`, status: http.StatusUnprocessableEntity, wantError: "out of the numeric range"},
		{name: "missing field a", method: http.MethodPost, path: "/api/operations/add", body: `{"b":3}`, status: http.StatusBadRequest, wantError: "field 'a' is required"},
		{name: "missing field b", method: http.MethodPost, path: "/api/operations/add", body: `{"a":2}`, status: http.StatusBadRequest, wantError: "field 'b' is required"},
		{name: "malformed JSON", method: http.MethodPost, path: "/api/operations/add", body: `{"a":2,`, status: http.StatusBadRequest, wantError: "invalid JSON body"},
		{name: "non-numeric field a", method: http.MethodPost, path: "/api/operations/add", body: `{"a":"2","b":3}`, status: http.StatusBadRequest, wantError: "invalid JSON body"},
		{name: "unknown field", method: http.MethodPost, path: "/api/operations/add", body: `{"a":2,"b":3,"c":4}`, status: http.StatusBadRequest, wantError: "invalid JSON body"},
		{name: "trailing data after the object", method: http.MethodPost, path: "/api/operations/add", body: `{"a":2,"b":3} junk`, status: http.StatusBadRequest, wantError: "invalid JSON body"},
		{name: "two JSON objects in one body", method: http.MethodPost, path: "/api/operations/add", body: `{"a":2,"b":3}{"a":1,"b":1}`, status: http.StatusBadRequest, wantError: "invalid JSON body"},
		{name: "overflow gives an out-of-range result", method: http.MethodPost, path: "/api/operations/multiply", body: `{"a":1e308,"b":10}`, status: http.StatusUnprocessableEntity, wantError: "out of the numeric range"},

		// Unary operations
		{name: "square root", method: http.MethodPost, path: "/api/operations/sqrt", body: `{"a":9}`, status: http.StatusOK, wantResult: floatPtr(3)},
		{name: "square root of a negative", method: http.MethodPost, path: "/api/operations/sqrt", body: `{"a":-9}`, status: http.StatusBadRequest, wantError: "square root of negative number"},
		{name: "square root without a", method: http.MethodPost, path: "/api/operations/sqrt", body: `{}`, status: http.StatusBadRequest, wantError: "field 'a' is required"},
		{name: "square root given b", method: http.MethodPost, path: "/api/operations/sqrt", body: `{"a":9,"b":2}`, status: http.StatusBadRequest, wantError: "invalid JSON body"},
		{name: "percent", method: http.MethodPost, path: "/api/operations/percent", body: `{"a":50}`, status: http.StatusOK, wantResult: floatPtr(0.5)},

		// Routing
		{name: "unknown operation", method: http.MethodPost, path: "/api/operations/logarithm", body: `{"a":2,"b":3}`, status: http.StatusNotFound, wantError: "unknown operation"},
		{name: "GET not allowed", method: http.MethodGet, path: "/api/operations/add", body: ``, status: http.StatusMethodNotAllowed},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			rec := httptest.NewRecorder()

			NewRouter().ServeHTTP(rec, req)

			if rec.Code != c.status {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, c.status, rec.Body.String())
			}

			if c.wantResult != nil {
				var body resultResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("body is not valid JSON: %v", err)
				}
				if !approxEqual(body.Result, *c.wantResult) {
					t.Errorf("result = %v, want %v", body.Result, *c.wantResult)
				}
			}

			if c.wantError != "" {
				var body errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("error body is not valid JSON: %v", err)
				}
				if !strings.Contains(body.Error, c.wantError) {
					t.Errorf("error = %q, want it to contain %q", body.Error, c.wantError)
				}
			}
		})
	}
}

// approxEqual compares floats with a tolerance. See the README for the reasoning.
func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func floatPtr(v float64) *float64 {
	return &v
}
