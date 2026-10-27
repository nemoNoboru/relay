export type AuthGroup = 'auth' | 'app'

/**
 * Where a route group must send the visitor, or null to render it. Signed-out visitors may only
 * see the (auth) group (login, join); signed-in visitors are moved out of it. While the stored
 * session is being verified nothing redirects, so a valid session never flashes the login screen.
 */
export function authRedirect({ group, signedIn, isLoading }: { group: AuthGroup; signedIn: boolean; isLoading: boolean }): '/' | '/login' | null {
  if (isLoading) return null
  if (group === 'app') return signedIn ? null : '/login'
  return signedIn ? '/' : null
}
