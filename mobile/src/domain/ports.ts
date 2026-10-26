import type { AuthError } from './auth-error'
import type { AuthTokens, Credentials, RefreshedSession, Session } from './session'
import type { Result } from './result'

export type AuthRepository = {
  readonly login: (credentials: Credentials) => Promise<Result<Session, AuthError>>
  readonly refresh: (refreshToken: string) => Promise<Result<RefreshedSession, AuthError>>
}

export type TokenReader = {
  /** Null when no access token is stored. */
  readonly load: () => Promise<AuthTokens | null>
}

export type TokenWriter = {
  /** An undefined refresh token leaves the stored one untouched. */
  readonly save: (tokens: AuthTokens) => Promise<void>
  readonly clear: () => Promise<void>
}

export type ServerUrlReader = {
  /** The stored server URL, or the default when none is stored. */
  readonly get: () => Promise<string>
}

export type ServerUrlWriter = {
  readonly set: (url: string) => Promise<void>
}

export type SessionEvents = {
  /** Fires when the session can no longer be refreshed. Returns an unsubscribe function. */
  readonly onExpired: (handler: () => void) => () => void
}
