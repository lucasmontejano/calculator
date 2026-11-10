package httpapi

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// Random tests with a fixed seed: the values vary, but are always the same,
// so a failure can always be reproduced exactly.
const randomCases = 300

func newGenerator() *rand.Rand {
	return rand.New(rand.NewPCG(20261108, 1))
}

// randomNumber returns a value between -1000 and 1000 with up to two decimal places.
func randomNumber(rng *rand.Rand) float64 {
	return math.Round((rng.Float64()*2000-1000)*100) / 100
}

// formatJSON writes the number without scientific notation, which is always valid JSON.
func formatJSON(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// nearlyEqual compares with a relative tolerance, because the values can be large.
func nearlyEqual(got, want float64) bool {
	return math.Abs(got-want) <= 1e-9*math.Max(1, math.Abs(want))
}

func send(operation, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/operations/"+operation, strings.NewReader(body))
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)
	return rec
}

func TestRandomBinaryOperations(t *testing.T) {
	// A slice, not a map, so the order is fixed and the generator always produces the same cases.
	operations := []struct {
		name string
		want func(a, b float64) float64
	}{
		{"add", func(a, b float64) float64 { return a + b }},
		{"subtract", func(a, b float64) float64 { return a - b }},
		{"multiply", func(a, b float64) float64 { return a * b }},
	}

	rng := newGenerator()
	for _, op := range operations {
		for i := 0; i < randomCases; i++ {
			a, b := randomNumber(rng), randomNumber(rng)
			body := fmt.Sprintf(`{"a":%s,"b":%s}`, formatJSON(a), formatJSON(b))

			rec := send(op.name, body)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s %s: status %d, want 200 (%s)", op.name, body, rec.Code, rec.Body.String())
			}

			var resp resultResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("%s %s: invalid body: %v", op.name, body, err)
			}
			if resp.Operation != op.name {
				t.Errorf("operation = %q, want %q", resp.Operation, op.name)
			}
			if want := op.want(a, b); !nearlyEqual(resp.Result, want) {
				t.Errorf("%s(%v, %v) = %v, want %v", op.name, a, b, resp.Result, want)
			}
		}
	}
}

func TestRandomDivisions(t *testing.T) {
	rng := newGenerator()
	for i := 0; i < randomCases; i++ {
		a := randomNumber(rng)
		b := randomNumber(rng)
		if i%10 == 0 {
			b = 0 // division by zero in 10% of the cases
		}
		body := fmt.Sprintf(`{"a":%s,"b":%s}`, formatJSON(a), formatJSON(b))

		rec := send("divide", body)
		if b == 0 {
			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "division by zero") {
				t.Errorf("divide %s: status %d, want 400 with division by zero", body, rec.Code)
			}
			continue
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("divide %s: status %d, want 200", body, rec.Code)
		}
		var resp resultResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("divide %s: invalid body: %v", body, err)
		}
		if !nearlyEqual(resp.Result, a/b) {
			t.Errorf("divide(%v, %v) = %v, want %v", a, b, resp.Result, a/b)
		}
	}
}

func TestRandomPowers(t *testing.T) {
	rng := newGenerator()
	for i := 0; i < randomCases; i++ {
		base := float64(rng.IntN(21) - 10)   // from -10 to 10
		exponent := float64(rng.IntN(9) - 3) // from -3 to 5
		body := fmt.Sprintf(`{"a":%s,"b":%s}`, formatJSON(base), formatJSON(exponent))

		expected := math.Pow(base, exponent)
		rec := send("power", body)

		// 0 raised to a negative exponent gives Inf, which the API rejects with 422.
		if math.IsInf(expected, 0) {
			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("power %s: status %d, want 422", body, rec.Code)
			}
			continue
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("power %s: status %d, want 200", body, rec.Code)
		}
		var resp resultResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("power %s: invalid body: %v", body, err)
		}
		if !nearlyEqual(resp.Result, expected) {
			t.Errorf("power(%v, %v) = %v, want %v", base, exponent, resp.Result, expected)
		}
	}
}

func TestRandomSqrtAndPercent(t *testing.T) {
	rng := newGenerator()
	for i := 0; i < randomCases; i++ {
		a := randomNumber(rng)
		body := fmt.Sprintf(`{"a":%s}`, formatJSON(a))

		rec := send("sqrt", body)
		if a < 0 {
			if rec.Code != http.StatusBadRequest {
				t.Errorf("sqrt %s: status %d, want 400", body, rec.Code)
			}
		} else {
			var resp resultResponse
			if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &resp) != nil {
				t.Fatalf("sqrt %s: status %d, want 200", body, rec.Code)
			}
			if !nearlyEqual(resp.Result, math.Sqrt(a)) {
				t.Errorf("sqrt(%v) = %v, want %v", a, resp.Result, math.Sqrt(a))
			}
		}

		rec = send("percent", body)
		var resp resultResponse
		if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &resp) != nil {
			t.Fatalf("percent %s: status %d, want 200", body, rec.Code)
		}
		if !nearlyEqual(resp.Result, a/100) {
			t.Errorf("percent(%v) = %v, want %v", a, resp.Result, a/100)
		}
	}
}
