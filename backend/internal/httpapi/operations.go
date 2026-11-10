package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"

	"calculator-app/backend/internal/calculator"
)

// maxBodyBytes limits the size of the request body.
const maxBodyBytes = 1 << 10 // 1 KiB

// binaryOperation takes two operands: body {"a":..., "b":...}.
type binaryOperation func(a, b float64) (float64, error)

// unaryOperation takes one operand: body {"a":...}.
type unaryOperation func(a float64) (float64, error)

var binaryOperations = map[string]binaryOperation{
	"add":      func(a, b float64) (float64, error) { return calculator.Add(a, b), nil },
	"subtract": func(a, b float64) (float64, error) { return calculator.Subtract(a, b), nil },
	"multiply": func(a, b float64) (float64, error) { return calculator.Multiply(a, b), nil },
	"divide":   calculator.Divide,
	"power":    func(a, b float64) (float64, error) { return calculator.Power(a, b), nil },
}

var unaryOperations = map[string]unaryOperation{
	"sqrt":    calculator.Sqrt,
	"percent": func(a float64) (float64, error) { return calculator.Percent(a), nil },
}

type binaryOperands struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

type unaryOperands struct {
	A *float64 `json:"a"`
}

// resultResponse is the success response. It echoes the operation and the operands to ease auditing
// and debugging. "b" only appears in binary operations.
type resultResponse struct {
	Operation string   `json:"operation"`
	A         float64  `json:"a"`
	B         *float64 `json:"b,omitempty"`
	Result    float64  `json:"result"`
}

// errorResponse is the format of every error: always an object with the "error" key.
type errorResponse struct {
	Error string `json:"error"`
}

func handleOperation(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("op")

	if op, ok := binaryOperations[name]; ok {
		handleBinary(w, r, name, op)
		return
	}
	if op, ok := unaryOperations[name]; ok {
		handleUnary(w, r, name, op)
		return
	}
	writeError(w, http.StatusNotFound, "unknown operation")
}

func handleBinary(w http.ResponseWriter, r *http.Request, name string, op binaryOperation) {
	var in binaryOperands
	if !decodeBody(w, r, &in) {
		return
	}
	if in.A == nil {
		writeError(w, http.StatusBadRequest, "field 'a' is required")
		return
	}
	if in.B == nil {
		writeError(w, http.StatusBadRequest, "field 'b' is required")
		return
	}

	result, err := op(*in.A, *in.B)
	respond(w, resultResponse{Operation: name, A: *in.A, B: in.B}, result, err)
}

func handleUnary(w http.ResponseWriter, r *http.Request, name string, op unaryOperation) {
	var in unaryOperands
	if !decodeBody(w, r, &in) {
		return
	}
	if in.A == nil {
		writeError(w, http.StatusBadRequest, "field 'a' is required")
		return
	}

	result, err := op(*in.A)
	respond(w, resultResponse{Operation: name, A: *in.A}, result, err)
}

// decodeBody reads the JSON body. Unknown fields are an error, so "b" in a unary operation is rejected.
// The body must hold exactly one JSON value, so anything after it is rejected too.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	// Token returns io.EOF only when nothing but whitespace is left after the object.
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON body: unexpected data after the object")
		return false
	}
	return true
}

// respond picks the HTTP status from the error or the result.
// base already carries the operation and the operands; the result is filled in here.
func respond(w http.ResponseWriter, base resultResponse, result float64, err error) {
	if errors.Is(err, calculator.ErrDivisionByZero) || errors.Is(err, calculator.ErrNegativeSqrt) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// JSON cannot represent NaN or Inf. A negative base with a fractional exponent gives NaN;
	// overflow, or zero raised to a negative exponent, gives Inf.
	if math.IsNaN(result) {
		writeError(w, http.StatusUnprocessableEntity, "result is not a real number")
		return
	}
	if math.IsInf(result, 0) {
		writeError(w, http.StatusUnprocessableEntity, "result is out of the numeric range")
		return
	}

	base.Result = result
	writeJSON(w, http.StatusOK, base)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
