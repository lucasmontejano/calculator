// Package calculator implements the arithmetic operations exposed by the API.
package calculator

import (
	"errors"
	"math"
)

// ErrDivisionByZero is returned when the divisor of an operation is zero.
var ErrDivisionByZero = errors.New("division by zero")

// ErrNegativeSqrt is returned when taking the square root of a negative number.
var ErrNegativeSqrt = errors.New("square root of negative number")

// Add returns a + b.
func Add(a, b float64) float64 {
	return a + b
}

// Subtract returns a - b.
func Subtract(a, b float64) float64 {
	return a - b
}

// Multiply returns a * b.
func Multiply(a, b float64) float64 {
	return a * b
}

// Divide returns a / b, or ErrDivisionByZero if b is zero.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Power returns a raised to the power of b. It may return NaN or Inf, which the HTTP layer handles.
func Power(a, b float64) float64 {
	return math.Pow(a, b)
}

// Sqrt returns the square root of a, or ErrNegativeSqrt if a is negative.
func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSqrt
	}
	return math.Sqrt(a), nil
}

// Percent returns a percent of a, that is, a / 100.
func Percent(a float64) float64 {
	return a / 100
}
