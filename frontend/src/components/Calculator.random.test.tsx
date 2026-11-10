import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { Calculator } from './Calculator'

// Pseudo-random generator with a fixed seed (mulberry32): the sequence is always the same,
// so a failure can be reproduced with the same seed.
function createGenerator(seed: number): () => number {
  let state = seed >>> 0
  return () => {
    state = (state + 0x6d2b79f5) >>> 0
    let t = state
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

// Operations used in the tests: the key, the API name and the reference calculation.
// Only integer-result operations, so results can be compared exactly.
const OPERATIONS = [
  { key: '+', name: 'add', calculate: (a: number, b: number) => a + b },
  { key: '-', name: 'subtract', calculate: (a: number, b: number) => a - b },
  { key: '*', name: 'multiply', calculate: (a: number, b: number) => a * b },
] as const

// Simulates the backend with the same JavaScript arithmetic.
function fakeFetchCalculator() {
  return vi.fn(async (url: string, init?: RequestInit) => {
    const name = url.split('/').pop() ?? ''
    const { a, b } = JSON.parse(String(init?.body)) as { a: number; b: number }
    const operation = OPERATIONS.find((o) => o.name === name)
    if (!operation) throw new Error(`unexpected operation: ${name}`)
    return new Response(JSON.stringify({ result: operation.calculate(a, b) }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })
  })
}

const display = () => screen.getByRole('status').textContent

describe('Calculator: random tests', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', fakeFetchCalculator())
  })

  afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
  })

  it('gets the right result for 60 random chains, evaluated left to right', async () => {
    const random = createGenerator(2026)
    const user = userEvent.setup()

    for (let caseIndex = 0; caseIndex < 60; caseIndex++) {
      const count = 2 + Math.floor(random() * 3) // from 2 to 4 numbers
      const numbers = Array.from({ length: count }, () => Math.floor(random() * 100))
      const operations = Array.from(
        { length: count - 1 },
        () => OPERATIONS[Math.floor(random() * OPERATIONS.length)],
      )

      // accumulated[i] is the value after applying the first i operations.
      const accumulated: number[] = [numbers[0]]
      operations.forEach((op, i) => accumulated.push(op.calculate(accumulated[i], numbers[i + 1])))

      const description = `case ${caseIndex}: ${numbers.map((n, i) => `${n} ${operations[i]?.key ?? '='}`).join(' ')}`
      const { unmount } = render(<Calculator />)

      for (let i = 0; i < count; i++) {
        await user.keyboard(String(numbers[i]))
        if (i < count - 1) {
          await user.keyboard(operations[i].key)
          // From the second operator on, the calculator already shows the partial result.
          if (i >= 1) {
            await waitFor(() => expect(display(), description).toBe(String(accumulated[i])))
          }
        }
      }

      await user.keyboard('=')
      await waitFor(() => expect(display(), description).toBe(String(accumulated[count - 1])))
      unmount()
    }
  })

  it('pressing "=" repeatedly reuses the same operation and second number', async () => {
    const random = createGenerator(7)
    const user = userEvent.setup()

    for (let caseIndex = 0; caseIndex < 20; caseIndex++) {
      const a = Math.floor(random() * 10)
      const b = Math.floor(random() * 10)
      const op = OPERATIONS[Math.floor(random() * OPERATIONS.length)]
      const description = `case ${caseIndex}: ${a} ${op.key} ${b} =`
      const { unmount } = render(<Calculator />)

      await user.keyboard(`${a}${op.key}${b}=`)
      let expected = op.calculate(a, b)
      await waitFor(() => expect(display(), description).toBe(String(expected)))

      for (let repeat = 0; repeat < 3; repeat++) {
        await user.keyboard('=')
        expected = op.calculate(expected, b)
        await waitFor(() => expect(display(), `${description} (repeat ${repeat + 1})`).toBe(String(expected)))
      }
      unmount()
    }
  })
})
