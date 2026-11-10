package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// decodeObject reads the body as a generic object, to check exactly which keys exist.
func decodeObject(t *testing.T, body string) map[string]any {
	t.Helper()
	var obj map[string]any
	if err := json.Unmarshal([]byte(body), &obj); err != nil {
		t.Fatalf("body is not a JSON object: %v (body: %s)", err, body)
	}
	return obj
}

func TestSuccessfulBinaryResponseHasAllFields(t *testing.T) {
	rec := send("add", `{"a":2,"b":3}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	obj := decodeObject(t, rec.Body.String())
	want := map[string]any{"operation": "add", "a": 2.0, "b": 3.0, "result": 5.0}
	for key, value := range want {
		if obj[key] != value {
			t.Errorf("field %q = %v, want %v", key, obj[key], value)
		}
	}
	if len(obj) != len(want) {
		t.Errorf("keys = %v, want exactly %v", obj, want)
	}
}

func TestSuccessfulUnaryResponseHasNoBField(t *testing.T) {
	rec := send("sqrt", `{"a":9}`)

	obj := decodeObject(t, rec.Body.String())
	if _, exists := obj["b"]; exists {
		t.Errorf("unary operation should not have field b: %v", obj)
	}
	if obj["operation"] != "sqrt" || obj["result"] != 3.0 {
		t.Errorf("response = %v, want operation=sqrt and result=3", obj)
	}
}

func TestErrorResponseHasASingleShape(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{"division by zero", "divide", `{"a":1,"b":0}`, http.StatusBadRequest},
		{"unknown operation", "logarithm", `{"a":1,"b":1}`, http.StatusNotFound},
		{"invalid JSON", "add", `{`, http.StatusBadRequest},
		{"result is not real", "power", `{"a":-8,"b":0.5}`, http.StatusUnprocessableEntity},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := send(c.path, c.body)

			if rec.Code != c.status {
				t.Fatalf("status = %d, want %d", rec.Code, c.status)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			obj := decodeObject(t, rec.Body.String())
			if _, ok := obj["error"].(string); !ok || len(obj) != 1 {
				t.Errorf("error should be an object with only the error key: %v", obj)
			}
		})
	}
}

func TestNegativeDecimalResultIsEncodedPlainly(t *testing.T) {
	// Ensures Go's encoder produces valid JSON for negative and decimal values too.
	rec := send("multiply", `{"a":-0.5,"b":0.25}`)
	if !strings.Contains(rec.Body.String(), `"result":-0.125`) {
		t.Errorf("body = %s, want it to contain result -0.125", rec.Body.String())
	}
}
