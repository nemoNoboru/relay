import { AxiosError, type AxiosAdapter, type InternalAxiosRequestConfig } from 'axios'

const mockAsyncStore = new Map<string, string>()
const mockSecureStore = new Map<string, string>()

jest.mock('@react-native-async-storage/async-storage', () => ({
  __esModule: true,
  default: {
    getItem: jest.fn(async (k: string) => mockAsyncStore.get(k) ?? null),
    setItem: jest.fn(async (k: string, v: string) => void mockAsyncStore.set(k, v)),
  },
}))

jest.mock('expo-secure-store', () => ({
  getItemAsync: jest.fn(async (k: string) => mockSecureStore.get(k) ?? null),
  setItemAsync: jest.fn(async (k: string, v: string) => void mockSecureStore.set(k, v)),
  deleteItemAsync: jest.fn(async (k: string) => void mockSecureStore.delete(k)),
}))

import { api, DEFAULT_URL, LOGIN_PATH, onAuthFailure, REFRESH_PATH, setServerUrl } from '../api'

type Handler = (config: InternalAxiosRequestConfig) => { status: number; data?: unknown }

let seen: InternalAxiosRequestConfig[]

function useServer(handler: Handler) {
  const adapter: AxiosAdapter = async (config) => {
    seen.push(config)
    const { status, data = {} } = handler(config)
    const response = { data, status, statusText: String(status), headers: {}, config }
    if (status >= 400) {
      throw new AxiosError('failed', 'ERR_BAD_REQUEST', config, null, response)
    }
    return response
  }
  api.defaults.adapter = adapter
}

const urlOf = (c: InternalAxiosRequestConfig) => `${c.baseURL}${c.url}`

beforeEach(() => {
  mockAsyncStore.clear()
  mockSecureStore.clear()
  seen = []
})

describe('base URL', () => {
  it('uses the default when none is stored', async () => {
    useServer(() => ({ status: 200 }))
    await api.get('/ping')
    expect(urlOf(seen[0]!)).toBe(`${DEFAULT_URL}/ping`)
  })

  it('uses the stored URL', async () => {
    await setServerUrl('https://friends.example')
    useServer(() => ({ status: 200 }))
    await api.get('/ping')
    expect(urlOf(seen[0]!)).toBe('https://friends.example/ping')
  })

  it('picks up a changed URL on the next request', async () => {
    useServer(() => ({ status: 200 }))
    await setServerUrl('https://one.example')
    await api.get('/ping')
    await setServerUrl('https://two.example')
    await api.get('/ping')
    expect(seen.map(urlOf)).toEqual(['https://one.example/ping', 'https://two.example/ping'])
  })
})

describe('auth', () => {
  it('attaches the Bearer token', async () => {
    mockSecureStore.set('access_token', 'abc')
    useServer(() => ({ status: 200 }))
    await api.get('/ping')
    expect(seen[0]!.headers.get('Authorization')).toBe('Bearer abc')
  })

  it('sends no Authorization header without a token', async () => {
    useServer(() => ({ status: 200 }))
    await api.get('/ping')
    expect(seen[0]!.headers.get('Authorization')).toBeFalsy()
  })

  it('refreshes once on 401 and retries', async () => {
    mockSecureStore.set('access_token', 'old')
    mockSecureStore.set('refresh_token', 'r1')
    useServer((c) => {
      if (c.url === REFRESH_PATH) return { status: 200, data: { access_token: 'new', refresh_token: 'r2' } }
      return c.headers.get('Authorization') === 'Bearer new' ? { status: 200, data: { ok: true } } : { status: 401 }
    })
    const res = await api.get('/ping')
    expect(res.data).toEqual({ ok: true })
    expect(seen.filter((c) => c.url === REFRESH_PATH)).toHaveLength(1)
    expect(seen).toHaveLength(3)
    expect(mockSecureStore.get('access_token')).toBe('new')
    expect(mockSecureStore.get('refresh_token')).toBe('r2')
  })

  it('shares one refresh between concurrent 401s', async () => {
    mockSecureStore.set('access_token', 'old')
    mockSecureStore.set('refresh_token', 'r1')
    useServer((c) => {
      if (c.url === REFRESH_PATH) return { status: 200, data: { access_token: 'new' } }
      return c.headers.get('Authorization') === 'Bearer new' ? { status: 200 } : { status: 401 }
    })
    await Promise.all([api.get('/a'), api.get('/b')])
    expect(seen.filter((c) => c.url === REFRESH_PATH)).toHaveLength(1)
  })

  it('clears tokens and calls onAuthFailure when refresh fails, without a second refresh', async () => {
    mockSecureStore.set('access_token', 'old')
    mockSecureStore.set('refresh_token', 'r1')
    const failed = jest.fn()
    const off = onAuthFailure(failed)
    useServer(() => ({ status: 401 }))
    await expect(api.get('/ping')).rejects.toBeInstanceOf(AxiosError)
    off()
    expect(seen.filter((c) => c.url === REFRESH_PATH)).toHaveLength(1)
    expect(seen).toHaveLength(2)
    expect(mockSecureStore.size).toBe(0)
    expect(failed).toHaveBeenCalledTimes(1)
  })

  it('does not refresh or call onAuthFailure on a 401 from a skipAuthRefresh request (login)', async () => {
    mockSecureStore.set('access_token', 'old')
    mockSecureStore.set('refresh_token', 'r1')
    const failed = jest.fn()
    const off = onAuthFailure(failed)
    useServer(() => ({ status: 401 }))
    await expect(api.post(LOGIN_PATH, {}, { skipAuthRefresh: true })).rejects.toBeInstanceOf(AxiosError)
    off()
    expect(seen).toHaveLength(1)
    expect(failed).not.toHaveBeenCalled()
    expect(mockSecureStore.get('access_token')).toBe('old')
  })
})
