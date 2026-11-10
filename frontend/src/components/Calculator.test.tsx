import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { Calculator } from './Calculator'

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

// Simulates the backend: it really calculates and returns the same errors as the Go API.
function fakeFetch() {
  return vi.fn(async (url: string, init?: RequestInit) => {
    const operation = url.split('/').pop() ?? ''
    const { a, b } = JSON.parse(String(init?.body)) as { a: number; b?: number }

    if (operation === 'sqrt') {
      if (a < 0) return jsonResponse(400, { error: 'square root of negative number' })
      return jsonResponse(200, { result: Math.sqrt(a) })
    }
    if (operation === 'percent') return jsonResponse(200, { result: a / 100 })
    if (operation === 'divide' && b === 0) {
      return jsonResponse(400, { error: 'division by zero' })
    }

    const calculators: Record<string, (x: number, y: number) => number> = {
      add: (x, y) => x + y,
      subtract: (x, y) => x - y,
      multiply: (x, y) => x * y,
      divide: (x, y) => x / y,
      power: (x, y) => x ** y,
    }
    return jsonResponse(200, { result: calculators[operation](a, b ?? 0) })
  })
}

const display = () => screen.getByRole('status').textContent

describe('Calculator', () => {
  let fetchMock: ReturnType<typeof fakeFetch>

  beforeEach(() => {
    fetchMock = fakeFetch()
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    cleanup()
    vi.unstubAllGlobals()
  })

  it('adds two numbers with the buttons and shows the operation above', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.click(screen.getByRole('button', { name: '7' }))
    await user.click(screen.getByRole('button', { name: 'Add' }))
    await user.click(screen.getByRole('button', { name: '3' }))
    await user.click(screen.getByRole('button', { name: 'Equals' }))

    await waitFor(() => expect(display()).toBe('10'))
    expect(screen.getByText('7 + 3 =')).toBeInTheDocument()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/operations/add',
      expect.objectContaining({ body: JSON.stringify({ a: 7, b: 3 }) }),
    )
  })

  it('pressing "=" again repeats the last operation', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('9*6=')
    await waitFor(() => expect(display()).toBe('54'))

    await user.keyboard('=')
    await waitFor(() => expect(display()).toBe('324'))
    expect(screen.getByText('54 × 6 =')).toBeInTheDocument()
  })

  it('accepts a whole expression from the keyboard', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('2*4=')

    await waitFor(() => expect(display()).toBe('8'))
  })

  it('accepts a comma as the decimal separator and shows the result with a comma', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('0,1+0,2=')

    await waitFor(() => expect(display()).toBe('0,3'))
  })

  it('chains operations from left to right', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('2+3*')
    await waitFor(() => expect(display()).toBe('5'))

    await user.keyboard('4=')
    await waitFor(() => expect(display()).toBe('20'))
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it('shows the division-by-zero message on the display', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('1/0=')

    await waitFor(() => expect(display()).toBe('Cannot divide by zero'))
  })

  it('calculates powers from the keyboard', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('2^10=')

    await waitFor(() => expect(display()).toBe('1024'))
  })

  it('calculates a square root and shows the expression', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('9r')

    await waitFor(() => expect(display()).toBe('3'))
    expect(screen.getByText('√(9)')).toBeInTheDocument()
  })

  it('shows an error for the square root of a negative number', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('4')
    await user.click(screen.getByRole('button', { name: 'Toggle sign' }))
    await user.keyboard('r')

    await waitFor(() => expect(display()).toBe('Square root of a negative number'))
  })

  it('calculates the percent of the current number', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('50%')

    await waitFor(() => expect(display()).toBe('0,5'))
    expect(screen.getByText('50%')).toBeInTheDocument()
  })

  it('Backspace deletes the last digit', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('123{Backspace}')

    expect(display()).toBe('12')
  })

  it('Escape clears the display', async () => {
    const user = userEvent.setup()
    render(<Calculator />)

    await user.keyboard('45{Escape}')

    expect(display()).toBe('0')
  })

  it('ignores Escape while a calculation is still running', async () => {
    const user = userEvent.setup()
    let respond: (response: Response) => void = () => {}
    fetchMock.mockImplementationOnce(
      () =>
        new Promise<Response>((resolve) => {
          respond = resolve
        }),
    )
    render(<Calculator />)

    await user.keyboard('2+3=')
    await user.keyboard('{Escape}')
    expect(display()).toBe('3')

    respond(jsonResponse(200, { result: 5 }))
    await waitFor(() => expect(display()).toBe('5'))
  })
})
