// Learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom'

// Mock Next.js router
jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}))

// Mock Next Auth
jest.mock('next-auth/react', () => ({
  useSession: jest.fn(() => ({
    data: null,
    status: 'unauthenticated',
  })),
  getSession: jest.fn(),
  signIn: jest.fn(),
  signOut: jest.fn(),
}))

// Mock fetch API
global.fetch = jest.fn()

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
})

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
}
global.localStorage = localStorageMock

// Mock ResizeObserver for Radix UI components (Popper positioning)
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

// Mock PointerEvent for Radix UI (not available in JSDOM by default)
global.PointerEvent = class PointerEvent extends Event {
  constructor(type, props) {
    super(type, props)
    this.button = props?.button ?? 0
    this.ctrlKey = props?.ctrlKey ?? false
    this.pointerType = props?.pointerType ?? 'mouse'
  }
}

// Mock HTMLElement methods used by Radix UI for pointer capture and scroll
window.HTMLElement.prototype.scrollIntoView = jest.fn()
window.HTMLElement.prototype.releasePointerCapture = jest.fn()
window.HTMLElement.prototype.hasPointerCapture = jest.fn()

// Set timezone to UTC for consistent snapshot testing
process.env.TZ = 'UTC'
