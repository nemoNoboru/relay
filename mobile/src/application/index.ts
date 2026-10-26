import type { SessionEvents } from '../domain'
import type { makeLogin } from './login'
import type { makeLogout } from './logout'
import type { makeRestoreSession } from './restore-session'

export { makeLogin, type LoginInput } from './login'
export { makeLogout } from './logout'
export { makeRefreshSession } from './refresh-session'
export { makeRestoreSession } from './restore-session'

/** Everything the presentation layer may call; built once by the composition root. */
export type UseCases = {
  readonly login: ReturnType<typeof makeLogin>
  readonly logout: ReturnType<typeof makeLogout>
  readonly restoreSession: ReturnType<typeof makeRestoreSession>
  readonly sessionEvents: SessionEvents
}
