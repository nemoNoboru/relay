import type { SessionEvents } from '../domain'
import type { makeLogin } from './login'
import type { makeRegister } from './register'
import type { makeLogout } from './logout'
import type { makeRestoreSession } from './restore-session'

export { makeLogin, type LoginInput } from './login'
export { makeLogout } from './logout'
export { makeRegister, type RegisterInput } from './register'
export { makeRefreshSession } from './refresh-session'
export { makeRestoreSession } from './restore-session'

/** Everything the presentation layer may call; built once by the composition root. */
export type UseCases = {
  readonly login: ReturnType<typeof makeLogin>
  readonly register: ReturnType<typeof makeRegister>
  readonly logout: ReturnType<typeof makeLogout>
  readonly restoreSession: ReturnType<typeof makeRestoreSession>
  readonly sessionEvents: SessionEvents
}
