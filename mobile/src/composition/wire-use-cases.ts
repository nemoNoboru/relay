import type { AxiosAdapter } from 'axios'

import { makeLogin, makeLogout, makeRefreshSession, makeRestoreSession, type UseCases } from '../application'
import { makeHttpAuthRepository } from '../infrastructure/http/auth-repository'
import { makeHttpClient } from '../infrastructure/http/client'
import { makeSessionEvents } from '../infrastructure/session-events'
import type { KeyValueStorage } from '../infrastructure/storage/key-value'
import { makeServerUrlStore } from '../infrastructure/storage/server-url-store'
import { makeTokenStore } from '../infrastructure/storage/token-store'

export type WireDeps = {
  readonly secureStorage: KeyValueStorage
  readonly preferenceStorage: KeyValueStorage
  /** Replaces axios' network adapter; tests use it to fake the server. */
  readonly httpAdapter?: AxiosAdapter
}

/** The one place that knows which adapters implement which ports. */
export function wireUseCases({ secureStorage, preferenceStorage, httpAdapter }: WireDeps): UseCases {
  const tokenStore = makeTokenStore({ storage: secureStorage })
  const serverUrlStore = makeServerUrlStore({ storage: preferenceStorage })
  const sessionEvents = makeSessionEvents()

  // The client needs refreshSession and the repository (behind refreshSession) needs the client;
  // the closures below are only called once a request is made, after everything is wired.
  const http = makeHttpClient({
    getBaseUrl: serverUrlStore.get,
    getAccessToken: async () => (await tokenStore.load())?.accessToken ?? null,
    refreshAccessToken: async () => {
      const result = await refreshSession()
      return result.ok ? result.value.tokens.accessToken : null
    },
    onUnauthorized: async () => {
      await logout()
      sessionEvents.notifyExpired()
    },
  })
  if (httpAdapter) http.defaults.adapter = httpAdapter

  const authRepository = makeHttpAuthRepository({ http })
  const refreshSession = makeRefreshSession({ authRepository, tokenReader: tokenStore, tokenWriter: tokenStore })
  const logout = makeLogout({ tokenWriter: tokenStore })

  return {
    login: makeLogin({ authRepository, tokenWriter: tokenStore, serverUrlWriter: serverUrlStore }),
    logout,
    restoreSession: makeRestoreSession({ tokenReader: tokenStore, tokenWriter: tokenStore, refreshSession }),
    sessionEvents,
  }
}
