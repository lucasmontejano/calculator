package calculator

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

const randomCases = 500

func newGenerator() *rand.Rand {
	return rand.New(rand.NewPCG(20261108, 2))
}

func randomNumber(rng *rand.Rand) float64 {
	return math.Round((rng.Float64()*2000-1000)*100) / 100
}

// nearlyEqual compares with a relative tolerance. See the README for the reasoning.
func nearlyEqual(got, want float64) bool {
	return math.Abs(got-want) <= 1e-9*math.Max(1, math.Abs(want))
}

// Properties that must hold for any pair of inputs, not just hand-picked values.
func TestRandomProperties(t *testing.T) {
	rng := newGenerator()
	for i := 0; i < randomCases; i++ {
		a, b := randomNumber(rng), randomNumber(rng)

		if Add(a, b) != Add(b, a) {
			t.Fatalf("addition is not commutative for %v and %v", a, b)
		}
		if !nearlyEqual(Subtract(Add(a, b), b), a) {
			t.Errorf("(a + b) - b != a for a=%v, b=%v", a, b)
		}
		if Multiply(a, 0) != 0 {
			t.Errorf("a * 0 != 0 for a=%v", a)
		}
		if b != 0 {
			q, err := Divide(a, b)
			if err != nil {
				t.Fatalf("Divide(%v, %v) returned an unexpected error: %v", a, b, err)
			}
			if !nearlyEqual(Multiply(q, b), a) {
				t.Errorf("(a / b) * b != a for a=%v, b=%v", a, b)
			}
		}
	}
}

func TestSqrtOfRandomSquares(t *testing.T) {
	rng := newGenerator()
	for i := 0; i < randomCases; i++ {
		a := math.Abs(randomNumber(rng))
		r, err := Sqrt(a * a)
		if err != nil {
			t.Fatalf("Sqrt(%v) returned an unexpected error: %v", a*a, err)
		}
		if !nearlyEqual(r, a) {
			t.Errorf("Sqrt(%v²) = %v, want %v", a, r, a)
		}
	}
}

func TestSqrtOfRandomNegativesAlwaysFails(t *testing.T) {
	rng := newGenerator()
	for i := 0; i < randomCases; i++ {
		a := -math.Abs(randomNumber(rng)) - 1 // always less than zero
		if _, err := Sqrt(a); !errors.Is(err, ErrNegativeSqrt) {
			t.Errorf("Sqrt(%v) error = %v, want ErrNegativeSqrt", a, err)
		}
	}
}
