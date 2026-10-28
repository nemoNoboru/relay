import type { KeyValueStorage } from './key-value'

// expo-secure-store has no web implementation. localStorage is readable by any script on the
// origin, which is the best a browser offers; the app ships no third-party scripts.
export const makeSecureStoreStorage = (): KeyValueStorage => ({
  get: async (key) => window.localStorage.getItem(key),
  set: async (key, value) => window.localStorage.setItem(key, value),
  remove: async (key) => window.localStorage.removeItem(key),
})
