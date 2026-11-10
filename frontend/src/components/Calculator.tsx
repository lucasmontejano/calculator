import { useEffect, useState } from 'react'
import {
  ApiError,
  calculate,
  calculateUnary,
  type BinaryOperation,
  type UnaryOperation,
} from '../api/calculator'
import './Calculator.css'

const MAX_DIGITS = 15

const SYMBOL: Record<BinaryOperation, string> = {
  add: '+',
  subtract: '−',
  multiply: '×',
  divide: '÷',
  power: '^',
}

const ERROR_MESSAGES: Record<string, string> = {
  'division by zero': 'Cannot divide by zero',
  'square root of negative number': 'Square root of a negative number',
  'result is out of the numeric range': 'Result is too large',
  'result is not a real number': 'Result is not a real number',
}

// Physical keyboard keys that map to binary operations.
const OPERATION_BY_KEY: Record<string, BinaryOperation> = {
  '+': 'add',
  '-': 'subtract',
  '*': 'multiply',
  '/': 'divide',
  '^': 'power',
}

type State = {
  input: string // number being edited, or the last result, with a dot (e.g. "12.5")
  previous: number | null // first operand of the pending operation
  operation: BinaryOperation | null
  startsNewInput: boolean // the next digit starts a new number
  error: string | null
  expression: string // small line above the display, e.g. "9 × 6 ="
  lastCalc: { operation: BinaryOperation; operand: number } | null // used by repeated "="
}

const INITIAL_STATE: State = {
  input: '0',
  previous: null,
  operation: null,
  startsNewInput: true,
  error: null,
  expression: '',
  lastCalc: null,
}

function currentNumber(state: State): number {
  return Number(state.input)
}

// Shows numbers with a decimal comma, as is customary in Brazil.
function toDisplay(value: number | string): string {
  return String(value).replace('.', ',')
}

// Limits to 12 significant digits to hide the binary noise of float64.
function formatResult(value: number): string {
  return Number(value.toPrecision(12)).toString()
}

function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 0) return error.message
    return ERROR_MESSAGES[error.message] ?? 'Error'
  }
  return 'Error'
}

export function Calculator() {
  const [state, setState] = useState<State>(INITIAL_STATE)
  const [busy, setBusy] = useState(false)

  // Runs an API call and builds the new state from its result.
  async function run(build: (value: number) => State, request: () => Promise<number>) {
    setBusy(true)
    try {
      const value = await request()
      setState(build(value))
    } catch (error) {
      setState({ ...INITIAL_STATE, error: errorMessage(error) })
    } finally {
      setBusy(false)
    }
  }

  function enterDigit(digit: string) {
    if (busy) return
    setState((current) => {
      const base = current.error ? INITIAL_STATE : current
      if (base.startsNewInput) return { ...base, input: digit, startsNewInput: false }
      if (/^-?0$/.test(base.input)) return { ...base, input: base.input.replace('0', digit) }
      if (base.input.replace(/[-.]/g, '').length >= MAX_DIGITS) return current
      return { ...base, input: base.input + digit }
    })
  }

  function enterDecimal() {
    if (busy) return
    setState((current) => {
      const base = current.error ? INITIAL_STATE : current
      if (base.startsNewInput) return { ...base, input: '0.', startsNewInput: false }
      if (base.input.includes('.')) return current
      return { ...base, input: base.input + '.' }
    })
  }

  function toggleSign() {
    if (busy) return
    setState((current) => {
      if (current.error || current.input === '0') return current
      const input = current.input.startsWith('-') ? current.input.slice(1) : '-' + current.input
      return { ...current, input }
    })
  }

  function backspace() {
    if (busy) return
    setState((current) => {
      if (current.error || current.startsNewInput) return current
      const shorter = current.input.slice(0, -1)
      return { ...current, input: shorter === '' || shorter === '-' ? '0' : shorter }
    })
  }

  function clear() {
    if (busy) return
    setState(INITIAL_STATE)
  }

  async function pressOperation(operation: BinaryOperation) {
    if (busy) return
    const snapshot = state
    const pending = snapshot.operation
    const first = snapshot.previous
    const current = currentNumber(snapshot)

    // No pending operation: the current number becomes the first operand.
    if (pending === null || first === null) {
      setState({
        ...snapshot,
        error: null,
        previous: current,
        operation,
        startsNewInput: true,
        expression: `${toDisplay(snapshot.input)} ${SYMBOL[operation]}`,
        lastCalc: null,
      })
      return
    }

    // Operator pressed twice in a row: just replaces the operation.
    if (snapshot.startsNewInput) {
      setState({ ...snapshot, operation, expression: `${toDisplay(first)} ${SYMBOL[operation]}` })
      return
    }

    // There is a pending operation with a typed second operand: calculate before chaining.
    await run(
      (value) => ({
        ...INITIAL_STATE,
        input: formatResult(value),
        previous: value,
        operation,
        startsNewInput: true,
        expression: `${toDisplay(formatResult(value))} ${SYMBOL[operation]}`,
      }),
      () => calculate(pending, first, current),
    )
  }

  async function equals() {
    if (busy || state.error) return
    const snapshot = state

    // Pending operation: calculate, and remember the operator and second operand for the next "=".
    if (snapshot.operation !== null && snapshot.previous !== null) {
      const operation = snapshot.operation
      const first = snapshot.previous
      const second = currentNumber(snapshot)
      await run(
        (value) => ({
          ...INITIAL_STATE,
          input: formatResult(value),
          expression: `${toDisplay(first)} ${SYMBOL[operation]} ${toDisplay(second)} =`,
          lastCalc: { operation, operand: second },
        }),
        () => calculate(operation, first, second),
      )
      return
    }

    // No pending operation: repeat the last one, using the current number as the first operand.
    if (snapshot.lastCalc !== null) {
      const { operation, operand } = snapshot.lastCalc
      const first = currentNumber(snapshot)
      await run(
        (value) => ({
          ...INITIAL_STATE,
          input: formatResult(value),
          expression: `${toDisplay(first)} ${SYMBOL[operation]} ${toDisplay(operand)} =`,
          lastCalc: snapshot.lastCalc,
        }),
        () => calculate(operation, first, operand),
      )
    }
  }

  // Square root and percent: replace the current number, without clearing the pending operation.
  async function pressUnary(operation: UnaryOperation) {
    if (busy || state.error) return
    const snapshot = state
    const current = currentNumber(snapshot)
    const fragment =
      operation === 'sqrt' ? `√(${toDisplay(snapshot.input)})` : `${toDisplay(snapshot.input)}%`
    const prefix =
      snapshot.operation !== null && snapshot.previous !== null
        ? `${toDisplay(snapshot.previous)} ${SYMBOL[snapshot.operation]} `
        : ''

    await run(
      (value) => ({
        ...snapshot,
        input: formatResult(value),
        startsNewInput: true,
        expression: prefix + fragment,
      }),
      () => calculateUnary(operation, current),
    )
  }

  function actionForKey(key: string): (() => void) | null {
    if (/^[0-9]$/.test(key)) return () => enterDigit(key)
    if (key === '.' || key === ',') return enterDecimal
    if (key === 'Enter' || key === '=') return () => void equals()
    if (key === 'Backspace') return backspace
    if (key === 'Escape' || key === 'Delete') return clear
    if (key === '%') return () => void pressUnary('percent')
    if (key.toLowerCase() === 'r') return () => void pressUnary('sqrt')
    const operation = OPERATION_BY_KEY[key]
    if (operation) return () => void pressOperation(operation)
    return null
  }

  // No dependency array: the listener always uses the handlers from the current render.
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      const action = actionForKey(event.key)
      if (action) {
        event.preventDefault()
        action()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  })

  const shown = state.error ?? toDisplay(state.input)

  return (
    <main className="calculator">
      <div className="expression">{state.expression}</div>
      <hr className="divider" />
      <output className={state.error ? 'display error' : 'display'}>{shown}</output>

      <div className="keypad">
        <button className="function-key" onClick={() => pressUnary('sqrt')} aria-label="Square root">√</button>
        <button className="function-key" onClick={() => pressOperation('power')} aria-label="Power">^</button>
        <button className="function-key" onClick={() => pressUnary('percent')} aria-label="Percent">%</button>
        <button className="operator" onClick={() => pressOperation('divide')} aria-label="Divide">÷</button>

        <button className="function-key" onClick={clear}>C</button>
        <button className="function-key" onClick={backspace} aria-label="Backspace">⌫</button>
        <button className="function-key" onClick={toggleSign} aria-label="Toggle sign">±</button>
        <button className="operator" onClick={() => pressOperation('multiply')} aria-label="Multiply">×</button>

        <button onClick={() => enterDigit('7')}>7</button>
        <button onClick={() => enterDigit('8')}>8</button>
        <button onClick={() => enterDigit('9')}>9</button>
        <button className="operator" onClick={() => pressOperation('subtract')} aria-label="Subtract">−</button>

        <button onClick={() => enterDigit('4')}>4</button>
        <button onClick={() => enterDigit('5')}>5</button>
        <button onClick={() => enterDigit('6')}>6</button>
        <button className="operator" onClick={() => pressOperation('add')} aria-label="Add">+</button>

        <button onClick={() => enterDigit('1')}>1</button>
        <button onClick={() => enterDigit('2')}>2</button>
        <button onClick={() => enterDigit('3')}>3</button>
        <button className="equals" onClick={() => equals()} aria-label="Equals">=</button>

        <button className="zero" onClick={() => enterDigit('0')}>0</button>
        <button onClick={enterDecimal} aria-label="Decimal separator">,</button>
      </div>
    </main>
  )
}
