import { DEFAULT_SERVER_URL, type ServerUrlReader, type ServerUrlWriter } from '../../domain'
import type { KeyValueStorage } from './key-value'

export const SERVER_URL_KEY = 'relay_server_url'

export const makeServerUrlStore = ({ storage }: { storage: KeyValueStorage }): ServerUrlReader & ServerUrlWriter => ({
  async get() {
    return (await storage.get(SERVER_URL_KEY)) ?? DEFAULT_SERVER_URL
  },

  set: (url) => storage.set(SERVER_URL_KEY, url),
})
