import {
  err,
  normalizeServerUrl,
  ok,
  type AuthError,
  type AuthRepository,
  type Result,
  type ServerUrlWriter,
  type TokenWriter,
  type User,
} from '../domain'

export type LoginInput = {
  readonly serverUrl: string
  readonly username: string
  readonly password: string
}

type Deps = {
  readonly authRepository: Pick<AuthRepository, 'login'>
  readonly tokenWriter: Pick<TokenWriter, 'save'>
  readonly serverUrlWriter: ServerUrlWriter
}

export const makeLogin =
  ({ authRepository, tokenWriter, serverUrlWriter }: Deps) =>
  async ({ serverUrl, username, password }: LoginInput): Promise<Result<User, AuthError>> => {
    // The server must be stored first: the login request reads it on every call.
    if (serverUrl.trim()) await serverUrlWriter.set(normalizeServerUrl(serverUrl))
    const result = await authRepository.login({ username, password })
    if (!result.ok) return err(result.error)
    await tokenWriter.save(result.value.tokens)
    return ok(result.value.user)
  }
