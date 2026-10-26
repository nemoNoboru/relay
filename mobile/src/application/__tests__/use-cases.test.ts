import { err, ok, type RefreshedSession, type Session } from '../../domain'
import { makeFakeAuthRepository, makeMemoryServerUrlStore, makeMemoryTokenStore } from '../../testing/fakes'
import { makeLogin, makeLogout, makeRegister, makeRefreshSession, makeRestoreSession } from '..'

const user = { id: 'u1', name: 'Ada' }
const session: Session = { user, tokens: { accessToken: 'a1', refreshToken: 'r1' } }
const refreshed: RefreshedSession = { user, tokens: { accessToken: 'a2', refreshToken: 'r2' } }
const credentials = { username: 'ada', password: 'pw' }

describe('login', () => {
  it('stores the normalized server URL and the tokens, and returns the user', async () => {
    const { repository, calls } = makeFakeAuthRepository({ login: ok(session) })
    const tokens = makeMemoryTokenStore()
    const serverUrl = makeMemoryServerUrlStore()
    const login = makeLogin({ authRepository: repository, tokenWriter: tokens, serverUrlWriter: serverUrl })

    const result = await login({ serverUrl: ' https://friends.example/ ', ...credentials })

    expect(result).toEqual(ok(user))
    expect(await serverUrl.get()).toBe('https://friends.example')
    expect(tokens.current()).toEqual(session.tokens)
    expect(calls.login).toEqual([credentials])
  })

  it('keeps the stored server URL when the input is blank', async () => {
    const { repository } = makeFakeAuthRepository({ login: ok(session) })
    const serverUrl = makeMemoryServerUrlStore('https://kept.example')
    const login = makeLogin({ authRepository: repository, tokenWriter: makeMemoryTokenStore(), serverUrlWriter: serverUrl })
    await login({ serverUrl: '  ', ...credentials })
    expect(await serverUrl.get()).toBe('https://kept.example')
  })

  it('returns the error and stores no tokens when the repository fails', async () => {
    const { repository } = makeFakeAuthRepository({ login: err('invalid-credentials') })
    const tokens = makeMemoryTokenStore()
    const login = makeLogin({ authRepository: repository, tokenWriter: tokens, serverUrlWriter: makeMemoryServerUrlStore() })
    expect(await login({ serverUrl: '', ...credentials })).toEqual(err('invalid-credentials'))
    expect(tokens.current()).toBeNull()
  })
})

describe('register', () => {
  const input = { inviteToken: 'inv', displayName: ' Ada L ', username: ' ada ', password: 'pw' }

  it('sends the trimmed fields, stores the tokens and returns the user', async () => {
    const { repository, calls } = makeFakeAuthRepository({ register: ok(session) })
    const tokens = makeMemoryTokenStore()
    const result = await makeRegister({ authRepository: repository, tokenWriter: tokens })(input)
    expect(result).toEqual(ok(user))
    expect(tokens.current()).toEqual(session.tokens)
    expect(calls.register).toEqual([{ inviteToken: 'inv', displayName: 'Ada L', username: 'ada', password: 'pw' }])
  })

  it.each(['invalid-invite', 'username-taken'] as const)('returns %s and stores no tokens', async (error) => {
    const { repository } = makeFakeAuthRepository({ register: err(error) })
    const tokens = makeMemoryTokenStore()
    expect(await makeRegister({ authRepository: repository, tokenWriter: tokens })(input)).toEqual(err(error))
    expect(tokens.current()).toBeNull()
  })
})

describe('logout', () => {
  it('clears the stored tokens', async () => {
    const tokens = makeMemoryTokenStore({ accessToken: 'a', refreshToken: 'r' })
    await makeLogout({ tokenWriter: tokens })()
    expect(tokens.current()).toBeNull()
  })
})

describe('refreshSession', () => {
  it('sends the stored refresh token and saves the rotated tokens', async () => {
    const { repository, calls } = makeFakeAuthRepository({ refresh: ok(refreshed) })
    const tokens = makeMemoryTokenStore({ accessToken: 'a1', refreshToken: 'r1' })
    const refresh = makeRefreshSession({ authRepository: repository, tokenReader: tokens, tokenWriter: tokens })

    expect(await refresh()).toEqual(ok(refreshed))
    expect(calls.refresh).toEqual(['r1'])
    expect(tokens.current()).toEqual(refreshed.tokens)
  })

  it('keeps the old refresh token when the response has none', async () => {
    const { repository } = makeFakeAuthRepository({ refresh: ok({ user: null, tokens: { accessToken: 'a2' } }) })
    const tokens = makeMemoryTokenStore({ accessToken: 'a1', refreshToken: 'r1' })
    await makeRefreshSession({ authRepository: repository, tokenReader: tokens, tokenWriter: tokens })()
    expect(tokens.current()).toEqual({ accessToken: 'a2', refreshToken: 'r1' })
  })

  it('fails with no-session without calling the server when there is no refresh token', async () => {
    const { repository, calls } = makeFakeAuthRepository({ refresh: ok(refreshed) })
    const tokens = makeMemoryTokenStore({ accessToken: 'a1' })
    const refresh = makeRefreshSession({ authRepository: repository, tokenReader: tokens, tokenWriter: tokens })
    expect(await refresh()).toEqual(err('no-session'))
    expect(calls.refresh).toHaveLength(0)
  })

  it('shares one request between concurrent callers, then allows a new one', async () => {
    let release: () => void = () => {}
    const gate = new Promise<void>((resolve) => (release = resolve))
    const { repository, calls } = makeFakeAuthRepository({ refresh: async () => (await gate, ok(refreshed)) })
    const tokens = makeMemoryTokenStore({ accessToken: 'a1', refreshToken: 'r1' })
    const refresh = makeRefreshSession({ authRepository: repository, tokenReader: tokens, tokenWriter: tokens })

    const both = Promise.all([refresh(), refresh()])
    release()
    await both
    expect(calls.refresh).toHaveLength(1)

    await refresh()
    expect(calls.refresh).toHaveLength(2)
  })

  it('leaves the stored tokens alone when the refresh fails', async () => {
    const { repository } = makeFakeAuthRepository({ refresh: err('invalid-credentials') })
    const tokens = makeMemoryTokenStore({ accessToken: 'a1', refreshToken: 'r1' })
    const result = await makeRefreshSession({ authRepository: repository, tokenReader: tokens, tokenWriter: tokens })()
    expect(result).toEqual(err('invalid-credentials'))
    expect(tokens.current()).toEqual({ accessToken: 'a1', refreshToken: 'r1' })
  })
})

describe('restoreSession', () => {
  const setup = (stored: Parameters<typeof makeMemoryTokenStore>[0], refresh: Parameters<typeof makeFakeAuthRepository>[0]['refresh']) => {
    const { repository, calls } = makeFakeAuthRepository({ refresh })
    const tokens = makeMemoryTokenStore(stored)
    const refreshSession = makeRefreshSession({ authRepository: repository, tokenReader: tokens, tokenWriter: tokens })
    const restore = makeRestoreSession({ tokenReader: tokens, tokenWriter: tokens, refreshSession })
    return { restore, tokens, calls }
  }

  it('returns null without a network call when nothing is stored', async () => {
    const { restore, calls } = setup(null, ok(refreshed))
    expect(await restore()).toBeNull()
    expect(calls.refresh).toHaveLength(0)
  })

  it('returns the user from a successful refresh', async () => {
    const { restore } = setup({ accessToken: 'a1', refreshToken: 'r1' }, ok(refreshed))
    expect(await restore()).toEqual(user)
  })

  it('returns null but keeps the tokens when the refresh response carries no user', async () => {
    const { restore, tokens } = setup({ accessToken: 'a1', refreshToken: 'r1' }, ok({ user: null, tokens: { accessToken: 'a2' } }))
    expect(await restore()).toBeNull()
    expect(tokens.current()).toEqual({ accessToken: 'a2', refreshToken: 'r1' })
  })

  it('clears stale tokens when the refresh fails', async () => {
    const { restore, tokens } = setup({ accessToken: 'a1', refreshToken: 'r1' }, err('invalid-credentials'))
    expect(await restore()).toBeNull()
    expect(tokens.current()).toBeNull()
  })
})
