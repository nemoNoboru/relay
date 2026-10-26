import AsyncStorage from '@react-native-async-storage/async-storage'

import type { KeyValueStorage } from './key-value'

export const makeAsyncStorage = (): KeyValueStorage => ({
  get: (key) => AsyncStorage.getItem(key),
  set: (key, value) => AsyncStorage.setItem(key, value),
  remove: (key) => AsyncStorage.removeItem(key),
})
