import type { AxiosAdapter } from 'axios'

import { makeLogin, makeLogout, makeRegister, makeRefreshSession, makeRestoreSession, type UseCases } from '../application'
import { makeHttpAuthRepository } from '../infrastructure/http/auth-repository'
import { makeHttpClient } from '../infrastructure/http/client'
import { makeSessionEvents } from '../infrastructure/session-events'
import type { KeyValueStorage } from '../infrastructure/storage/key-value'
import { makeServerUrlStore } from '../infrastructure/storage/server-url-store'
import { makeTokenStore } from '../infrastructure/storage/token-store'

const API_PREFIX = '/api'

export type WireDeps = {
  readonly secureStorage: KeyValueStorage
  readonly preferenceStorage: KeyValueStorage
  /** Replaces axios' network adapter; tests use it to fake the server. */
  readonly httpAdapter?: AxiosAdapter
  /** API base URL that ignores the stored server URL; the web build is served by its own backend. */
  readonly fixedApiBaseUrl?: string
}

/** The one place that knows which adapters implement which ports. */
export function wireUseCases({ secureStorage, preferenceStorage, httpAdapter, fixedApiBaseUrl }: WireDeps): UseCases {
  const tokenStore = makeTokenStore({ storage: secureStorage })
  const serverUrlStore = makeServerUrlStore({ storage: preferenceStorage })
  const sessionEvents = makeSessionEvents()

  // The client needs refreshSession and the repository (behind refreshSession) needs the client;
  // the closures below are only called once a request is made, after everything is wired.
  const http = makeHttpClient({
    getBaseUrl: async () => fixedApiBaseUrl ?? `${await serverUrlStore.get()}${API_PREFIX}`,
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
    register: makeRegister({ authRepository, tokenWriter: tokenStore }),
    logout,
    restoreSession: makeRestoreSession({ tokenReader: tokenStore, tokenWriter: tokenStore, refreshSession }),
    sessionEvents,
  }
}
