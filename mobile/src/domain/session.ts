import type { User } from './user'

export type AuthTokens = {
  readonly accessToken: string
  /** Servers may omit it on refresh, meaning "keep the current one". */
  readonly refreshToken?: string
}

export type Credentials = {
  readonly username: string
  readonly password: string
}

export type Registration = {
  readonly inviteToken: string
  readonly displayName: string
  readonly username: string
  readonly password: string
}

export type Session = {
  readonly user: User
  readonly tokens: AuthTokens
}

/** What a token refresh returns; the user is optional until the backend contract is fixed. */
export type RefreshedSession = {
  readonly user: User | null
  readonly tokens: AuthTokens
}

export const rotateTokens = (current: AuthTokens, next: AuthTokens): AuthTokens => ({
  accessToken: next.accessToken,
  refreshToken: next.refreshToken ?? current.refreshToken,
})
