import { authRedirect } from '../auth-guard'

describe('authRedirect', () => {
  it('sends a signed-out visitor from the app group to /login', () => {
    expect(authRedirect({ group: 'app', signedIn: false, isLoading: false })).toBe('/login')
  })

  it('lets a signed-in user into the app group', () => {
    expect(authRedirect({ group: 'app', signedIn: true, isLoading: false })).toBeNull()
  })

  it('lets a signed-out visitor see login and join', () => {
    expect(authRedirect({ group: 'auth', signedIn: false, isLoading: false })).toBeNull()
  })

  it('moves a signed-in user out of the auth group to the feed', () => {
    expect(authRedirect({ group: 'auth', signedIn: true, isLoading: false })).toBe('/')
  })

  it.each(['app', 'auth'] as const)('does not redirect %s while the session is being restored', (group) => {
    expect(authRedirect({ group, signedIn: false, isLoading: true })).toBeNull()
  })
})
