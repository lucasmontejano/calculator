# Calculator

A calculator that looks like the Windows standard calculator, built as a take-home exercise. The interface is React and TypeScript. The math runs on a small REST API written in Go, so every calculation goes through the backend, which checks the input and returns either a result or an error.

You can click the buttons or type on the keyboard.

## Running it

What you need: Go 1.25 or newer and Node.js 22 (Node 20.19 or newer also works). Docker is optional. There is no database and no environment file to set up.

### With Docker

```bash
docker compose up --build
```

Then open http://localhost:3000. The API listens on port 8080, and the Swagger UI is at http://localhost:8080/swagger.

### Without Docker

Run the backend and the frontend in two terminals.

```powershell
# terminal 1: backend, port 8080
cd backend
go run .\cmd\server

# terminal 2: frontend, port 5173
cd frontend
npm install
npm run dev
```

Open http://localhost:5173. The Vite dev server forwards `/api` to the Go server, so the browser only talks to one address.

To run the backend on another port, set `PORT` first. If you do, set `VITE_BACKEND_URL` to the same address before starting the frontend.

```powershell
$env:PORT = "9000"
go run .\cmd\server
```

On Linux or macOS, use `PORT=9000 go run ./cmd/server`.

## Using the calculator

| Key | Action |
|---|---|
| `0` to `9` | Type a digit |
| `.` or `,` | Decimal separator |
| `+` `-` `*` `/` | Add, subtract, multiply, divide |
| `^` | Power |
| `r` | Square root |
| `%` | Percent (the number divided by 100) |
| `Enter` or `=` | Calculate |
| `Backspace` | Delete the last digit |
| `Escape` or `Delete` | Clear |

The keypad uses the same actions: √ for square root, ^ for power, % for percent, ± to change the sign, ⌫ for backspace, and C to clear.

A few things that might not be obvious:

- Operations run from left to right with no precedence, so `2 + 3 × 4` gives `20`, the same as the Windows standard calculator.
- Pressing `=` again repeats the last operation with the number on the display. `9 × 6 =` gives `54`, and another `=` gives `324`.
- The display uses a decimal comma, which is the Brazilian convention. The keyboard accepts either `.` or `,`.
- Results are rounded to 12 significant digits. You can type up to 15 digits.
- While the server is working on a calculation, the other keys are ignored.
- After an error, type a new number or press C to start again.

## API

The API runs at `http://localhost:8080`. Each operation is a `POST` to `/api/operations/{op}` with a JSON body.

| Operation | Body | Result |
|---|---|---|
| `add` | `{"a": 2, "b": 3}` | a + b |
| `subtract` | `{"a": 2, "b": 3}` | a - b |
| `multiply` | `{"a": 2, "b": 3}` | a × b |
| `divide` | `{"a": 2, "b": 3}` | a ÷ b |
| `power` | `{"a": 2, "b": 10}` | a raised to the power b |
| `sqrt` | `{"a": 9}` | square root of a |
| `percent` | `{"a": 50}` | a / 100 |

`GET /api/health` returns `{"status":"ok"}`, which is handy for checking that the server is up.

Example request:

```bash
curl -X POST http://localhost:8080/api/operations/add \
  -H "Content-Type: application/json" \
  -d '{"a": 2, "b": 3}'
```

Response:

```json
{"operation":"add","a":2,"b":3,"result":5}
```

Unary operations take only `a`, and the response leaves out `b`.

On Windows, use `curl.exe`, because `curl` in PowerShell is an alias for `Invoke-WebRequest`. Escape the quotes like this:

```powershell
curl.exe -X POST http://localhost:8080/api/operations/divide -H "Content-Type: application/json" -d "{\"a\":1,\"b\":0}"
```

### Errors

Every error from the handlers has the same shape: an object with an `error` key.

```json
{"error":"division by zero"}
```

| Status | When | Example `error` |
|---|---|---|
| 400 | Division by zero, square root of a negative number, malformed JSON, a missing or unknown field, extra data after the JSON, or a body over 1 KiB | `division by zero` |
| 404 | The operation does not exist | `unknown operation` |
| 405 | Wrong method, such as a `GET` on an operation | `Method Not Allowed` (plain text) |
| 422 | The input is valid, but the result is not a real number or overflows | `result is not a real number` |
| 500 | Unexpected server error | `internal error` |

Requests to paths the router does not know get Go's plain-text 404, and wrong methods get the plain-text 405 shown above.

### Swagger

The interactive docs are at `http://localhost:8080/swagger`, and the raw OpenAPI spec is at `/swagger/openapi.yaml`. The Swagger page loads its assets from unpkg.com, so it needs internet access.

## Tests and coverage

Backend:

```powershell
cd backend
go test ./...
```

Frontend:

```powershell
cd frontend
npm test
```

The Go suite has 25 test functions with 53 subtests. The frontend has 15 tests.

To generate the coverage reports:

```powershell
# backend: writes docs\coverage\backend.html
cd backend
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o ..\docs\coverage\backend.html

# frontend: prints a summary and writes docs/coverage/frontend/index.html
cd frontend
npm run coverage
```

Numbers from the last run (statement coverage):

| Part | Coverage |
|---|---|
| Backend, `calculator` package | 100% |
| Backend, `httpapi` package | 97.2% |
| Backend, total | 90.4% (`cmd/server` has no tests because it only starts the server) |
| Frontend, `src/` | 78.7% |

## Project layout

```
.
├── backend/
│   ├── cmd/server/            entry point, reads PORT
│   ├── internal/calculator/   the arithmetic, no HTTP
│   ├── internal/httpapi/      routes, JSON, status codes, Swagger UI
│   └── Dockerfile
├── frontend/
│   ├── src/api/               client for the REST API
│   ├── src/components/        calculator UI and its tests
│   ├── nginx.conf             serves the build and forwards /api in Docker
│   └── Dockerfile
├── docs/
│   ├── openapi.yaml           API specification
│   └── coverage/              HTML coverage reports
└── docker-compose.yml
```

## How it is built

### Backend

The arithmetic lives in `internal/calculator`, and it knows nothing about HTTP. `internal/httpapi` handles routes, JSON and status codes. That split makes the math easy to test on its own.

The router is the standard library's `net/http`. Since Go 1.22, its `ServeMux` understands method and wildcard patterns such as `POST /api/operations/{op}`, so there is no router dependency.

Requests are strict. Unknown fields are rejected, the body must hold exactly one JSON value, and bodies over 1 KiB are refused. The operands are pointers, so a missing field is not mistaken for `0`.

Errors are sentinel values (`ErrDivisionByZero` and `ErrNegativeSqrt`). The HTTP layer picks the status with `errors.Is`, so rewording a message does not change behaviour. A 400 means the request itself is wrong. A 422 means the input was fine but the result cannot be used, such as a negative number raised to a fractional power.

Numbers are `float64`. Exact fractions cannot represent square roots or fractional powers, and the API needs both. The cost is that `0.1 + 0.2` gives `0.30000000000000004`. The tests compare with a tolerance, and the display rounds results to 12 significant digits.

The OpenAPI spec lives in `docs/openapi.yaml`. The binary embeds a copy from `internal/httpapi/static/`, because `go:embed` can only read files inside its own package. `TestOpenAPISpecMatchesDocs` fails when the two copies differ, so update both.

### Frontend

All of the calculator's state sits in one object, and each key press produces the next state from the previous one. While a request is pending, other keys are ignored, so two responses cannot arrive out of order and overwrite each other.

In development, Vite forwards `/api` to the Go server. The browser never makes a cross-origin request, so the backend needs no CORS setup. In Docker, nginx serves the built app and does the same forwarding.

### Tests

The unit tests are table-driven and use `t.Run`, so a failing case is named in the output. The random tests use fixed seeds, which means any failure can be reproduced exactly. Some tests check properties rather than single answers, such as addition being commutative and `(a / b) * b` giving back `a`. The frontend tests replace `fetch` with a stub that does the same arithmetic as the server, so they run without the backend.

## Limitations

- There is no authentication, rate limiting or history. Every request stands on its own.
- `float64` rounds very large or very precise values.
- `percent` is simply `a / 100`. The Windows calculator treats `200 + 10%` differently.
- Only basic left-to-right evaluation is implemented. There is no operator precedence.
- The tests run locally. There is no CI configuration.

## AI assistance

Parts of this project, including the code, the tests and this README, were written with help from Claude, an AI assistant made by Anthropic.
