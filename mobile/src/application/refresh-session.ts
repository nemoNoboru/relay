import {
  err,
  ok,
  rotateTokens,
  type AuthError,
  type AuthRepository,
  type RefreshedSession,
  type Result,
  type TokenReader,
  type TokenWriter,
} from '../domain'

type Deps = {
  readonly authRepository: Pick<AuthRepository, 'refresh'>
  readonly tokenReader: TokenReader
  readonly tokenWriter: Pick<TokenWriter, 'save'>
}

/**
 * Exchanges the stored refresh token for new tokens. Concurrent callers share one request, so
 * several 401s (or a 401 during launch restore) never spend the same refresh token twice.
 */
export const makeRefreshSession = ({ authRepository, tokenReader, tokenWriter }: Deps) => {
  let inFlight: Promise<Result<RefreshedSession, AuthError>> | null = null

  async function refresh(): Promise<Result<RefreshedSession, AuthError>> {
    const current = await tokenReader.load()
    if (!current?.refreshToken) return err('no-session')
    const result = await authRepository.refresh(current.refreshToken)
    if (!result.ok) return result
    const tokens = rotateTokens(current, result.value.tokens)
    await tokenWriter.save(tokens)
    return ok({ user: result.value.user, tokens })
  }

  return (): Promise<Result<RefreshedSession, AuthError>> => {
    inFlight ??= refresh().finally(() => {
      inFlight = null
    })
    return inFlight
  }
}
