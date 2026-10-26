import type { TokenReader, TokenWriter } from '../../domain'
import type { KeyValueStorage } from './key-value'

export const ACCESS_TOKEN_KEY = 'access_token'
export const REFRESH_TOKEN_KEY = 'refresh_token'

export const makeTokenStore = ({ storage }: { storage: KeyValueStorage }): TokenReader & TokenWriter => ({
  async load() {
    const accessToken = await storage.get(ACCESS_TOKEN_KEY)
    if (!accessToken) return null
    const refreshToken = await storage.get(REFRESH_TOKEN_KEY)
    return { accessToken, refreshToken: refreshToken ?? undefined }
  },

  async save({ accessToken, refreshToken }) {
    await storage.set(ACCESS_TOKEN_KEY, accessToken)
    if (refreshToken !== undefined) await storage.set(REFRESH_TOKEN_KEY, refreshToken)
  },

  async clear() {
    await storage.remove(ACCESS_TOKEN_KEY)
    await storage.remove(REFRESH_TOKEN_KEY)
  },
})
