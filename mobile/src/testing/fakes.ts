import {
  DEFAULT_SERVER_URL,
  err,
  type AuthError,
  type AuthRepository,
  type AuthTokens,
  type RefreshedSession,
  type Result,
  type ServerUrlReader,
  type ServerUrlWriter,
  type Session,
  type TokenReader,
  type TokenWriter,
} from '../domain'
import type { KeyValueStorage } from '../infrastructure/storage/key-value'

export const makeMemoryStorage = (initial: Record<string, string> = {}): KeyValueStorage & { data: Map<string, string> } => {
  const data = new Map(Object.entries(initial))
  return {
    data,
    get: async (key) => data.get(key) ?? null,
    set: async (key, value) => void data.set(key, value),
    remove: async (key) => void data.delete(key),
  }
}

export const makeMemoryTokenStore = (
  initial: AuthTokens | null = null,
): TokenReader & TokenWriter & { current: () => AuthTokens | null } => {
  let tokens = initial
  return {
    current: () => tokens,
    load: async () => tokens,
    save: async ({ accessToken, refreshToken }) => {
      tokens = { accessToken, refreshToken: refreshToken ?? tokens?.refreshToken }
    },
    clear: async () => {
      tokens = null
    },
  }
}

export const makeMemoryServerUrlStore = (initial = DEFAULT_SERVER_URL): ServerUrlReader & ServerUrlWriter => {
  let url = initial
  return { get: async () => url, set: async (next) => void (url = next) }
}

type Scripted<T> = Result<T, AuthError> | (() => Promise<Result<T, AuthError>>)

/** An AuthRepository answering from scripted results; records the calls it receives. */
export const makeFakeAuthRepository = (script: { login?: Scripted<Session>; refresh?: Scripted<RefreshedSession> }) => {
  const calls = { login: [] as unknown[], refresh: [] as string[] }
  const answer = <T>(scripted: Scripted<T> | undefined) =>
    typeof scripted === 'function' ? scripted() : Promise.resolve(scripted ?? err<AuthError>('unexpected'))
  const repository: AuthRepository = {
    login: (credentials) => (calls.login.push(credentials), answer(script.login)),
    refresh: (refreshToken) => (calls.refresh.push(refreshToken), answer(script.refresh)),
  }
  return { repository, calls }
}
