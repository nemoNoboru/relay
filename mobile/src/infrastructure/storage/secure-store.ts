import * as SecureStore from 'expo-secure-store'

import type { KeyValueStorage } from './key-value'

export const makeSecureStoreStorage = (): KeyValueStorage => ({
  get: (key) => SecureStore.getItemAsync(key),
  set: (key, value) => SecureStore.setItemAsync(key, value),
  remove: (key) => SecureStore.deleteItemAsync(key),
})
