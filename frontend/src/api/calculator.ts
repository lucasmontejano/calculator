export type BinaryOperation = 'add' | 'subtract' | 'multiply' | 'divide' | 'power'
export type UnaryOperation = 'sqrt' | 'percent'

export class ApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

type ResultBody = { result: number }
type ErrorBody = { error: string }

function isResultBody(value: unknown): value is ResultBody {
  return typeof value === 'object' && value !== null && 'result' in value && typeof value.result === 'number'
}

function isErrorBody(value: unknown): value is ErrorBody {
  return typeof value === 'object' && value !== null && 'error' in value && typeof value.error === 'string'
}

// Status 0 means the request never reached the server.
async function send(operation: string, body: { a: number; b?: number }): Promise<number> {
  let response: Response
  try {
    response = await fetch(`/api/operations/${operation}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
  } catch {
    throw new ApiError('Could not reach the server.', 0)
  }

  const data: unknown = await response.json().catch(() => null)

  if (!response.ok) {
    const message = isErrorBody(data) ? data.error : `HTTP ${response.status}`
    throw new ApiError(message, response.status)
  }

  if (!isResultBody(data)) {
    throw new ApiError('Invalid response from the server.', response.status)
  }

  return data.result
}

export function calculate(operation: BinaryOperation, a: number, b: number): Promise<number> {
  return send(operation, { a, b })
}

export function calculateUnary(operation: UnaryOperation, a: number): Promise<number> {
  return send(operation, { a })
}
