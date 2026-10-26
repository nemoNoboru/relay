import type { TokenReader, TokenWriter, User } from '../domain'
import type { makeRefreshSession } from './refresh-session'

type Deps = {
  readonly tokenReader: TokenReader
  readonly tokenWriter: Pick<TokenWriter, 'clear'>
  readonly refreshSession: ReturnType<typeof makeRefreshSession>
}

/**
 * On launch: a stored token is verified by refreshing it. Returns the signed-in user, or null
 * when there is no usable session (stale tokens are cleared).
 */
export const makeRestoreSession =
  ({ tokenReader, tokenWriter, refreshSession }: Deps) =>
  async (): Promise<User | null> => {
    if (!(await tokenReader.load())) return null
    const result = await refreshSession()
    if (result.ok) return result.value.user
    await tokenWriter.clear()
    return null
  }
