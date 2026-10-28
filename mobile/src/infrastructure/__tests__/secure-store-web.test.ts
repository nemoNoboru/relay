import { makeSecureStoreStorage } from '../storage/secure-store.web'

describe('web secure storage', () => {
  const data = new Map<string, string>()
  const localStorage = {
    getItem: (key: string) => data.get(key) ?? null,
    setItem: (key: string, value: string) => void data.set(key, value),
    removeItem: (key: string) => void data.delete(key),
  }

  beforeEach(() => {
    data.clear()
    Object.defineProperty(globalThis, 'window', { value: { localStorage }, configurable: true })
  })

  afterEach(() => {
    Reflect.deleteProperty(globalThis, 'window')
  })

  it('round-trips a value through localStorage', async () => {
    const storage = makeSecureStoreStorage()
    expect(await storage.get('k')).toBeNull()
    await storage.set('k', 'v')
    expect(await storage.get('k')).toBe('v')
    await storage.remove('k')
    expect(await storage.get('k')).toBeNull()
  })
})
