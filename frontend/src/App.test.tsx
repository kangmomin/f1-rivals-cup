import { render, screen } from '@testing-library/react'
import App from './App'

describe('App', () => {
  it('renders F1 Rivals Cup heading', () => {
    render(<App />)
    expect(screen.getByText('F1 Rivals Cup')).toBeInTheDocument()
  })

  it('renders with dark background', () => {
    render(<App />)
    const container = document.querySelector('.min-h-screen')
    expect(container).toHaveClass('bg-carbon')
  })

  it('displays correct subtitle', () => {
    render(<App />)
    expect(screen.getByText('League Management System')).toBeInTheDocument()
  })
})
