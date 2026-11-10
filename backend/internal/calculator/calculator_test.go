package calculator

import (
	"errors"
	"math"
	"testing"
)

// epsilon is the tolerance used to compare floats. See the README for the reasoning.
const epsilon = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positive integers", 2, 3, 5},
		{"with zero", 7, 0, 7},
		{"negatives", -4, -6, -10},
		{"positive and negative", 10, -3, 7},
		{"decimals with binary rounding error", 0.1, 0.2, 0.3},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Add(c.a, c.b); !approxEqual(got, c.want) {
				t.Errorf("Add(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	cases := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positive", 10, 4, 6},
		{"negative result", 4, 10, -6},
		{"with zero", 5, 0, 5},
		{"decimals", 1.5, 0.25, 1.25},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Subtract(c.a, c.b); !approxEqual(got, c.want) {
				t.Errorf("Subtract(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	cases := []struct {
		name string
		a, b float64
		want float64
	}{
		{"positives", 3, 4, 12},
		{"with zero", 9, 0, 0},
		{"two negatives", -3, -4, 12},
		{"positive and negative", -3, 4, -12},
		{"decimals", 0.5, 0.5, 0.25},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Multiply(c.a, c.b); !approxEqual(got, c.want) {
				t.Errorf("Multiply(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	cases := []struct {
		name string
		a, b float64
		want float64
	}{
		{"exact", 10, 2, 5},
		{"negative by positive", -10, 2, -5},
		{"two negatives", -10, -2, 5},
		{"decimal result", 1, 4, 0.25},
		{"repeating decimal", 1, 3, 0.333333333},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Divide(c.a, c.b)
			if err != nil {
				t.Fatalf("Divide(%v, %v) returned an unexpected error: %v", c.a, c.b, err)
			}
			if !approxEqual(got, c.want) {
				t.Errorf("Divide(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestDivideByZeroReturnsError(t *testing.T) {
	cases := []struct {
		name string
		a    float64
	}{
		{"positive by zero", 5},
		{"zero by zero", 0},
		{"negative by zero", -5},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Divide(c.a, 0)
			if !errors.Is(err, ErrDivisionByZero) {
				t.Fatalf("Divide(%v, 0) error = %v, want ErrDivisionByZero", c.a, err)
			}
			if got != 0 {
				t.Errorf("Divide(%v, 0) result = %v, want 0 on error", c.a, got)
			}
		})
	}
}

func TestPower(t *testing.T) {
	cases := []struct {
		name string
		a, b float64
		want float64
	}{
		{"integer", 2, 10, 1024},
		{"zero exponent", 7, 0, 1},
		{"negative exponent", 2, -2, 0.25},
		{"fractional exponent", 9, 0.5, 3},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Power(c.a, c.b); !approxEqual(got, c.want) {
				t.Errorf("Power(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	got, err := Sqrt(9)
	if err != nil {
		t.Fatalf("Sqrt(9) returned an unexpected error: %v", err)
	}
	if !approxEqual(got, 3) {
		t.Errorf("Sqrt(9) = %v, want 3", got)
	}

	if _, err := Sqrt(-1); !errors.Is(err, ErrNegativeSqrt) {
		t.Errorf("Sqrt(-1) error = %v, want ErrNegativeSqrt", err)
	}
}

func TestPercent(t *testing.T) {
	if got := Percent(50); !approxEqual(got, 0.5) {
		t.Errorf("Percent(50) = %v, want 0.5", got)
	}
}
