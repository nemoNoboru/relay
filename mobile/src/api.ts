import AsyncStorage from '@react-native-async-storage/async-storage'
import axios, { AxiosError } from 'axios'
import * as SecureStore from 'expo-secure-store'

export const BASE_URL_KEY = 'relay_server_url'
export const DEFAULT_URL = 'https://relay.mansus.org'
export const ACCESS_TOKEN_KEY = 'access_token'
export const REFRESH_TOKEN_KEY = 'refresh_token'

// The backend endpoints do not exist yet; keep paths here so they are easy to change.
export const LOGIN_PATH = '/auth/login'
export const REFRESH_PATH = '/auth/refresh'

export interface User {
  id: string
  name: string
}

export interface TokenPair {
  access_token: string
  refresh_token?: string
}

export interface LoginResponse extends TokenPair {
  refresh_token: string
  user: User
}

declare module 'axios' {
  interface AxiosRequestConfig {
    /** Do not refresh tokens or trigger onAuthFailure when this request gets a 401. */
    skipAuthRefresh?: boolean
    /** Set by the response interceptor once the request has been retried. */
    authRetried?: boolean
  }
}

export interface RefreshResponse extends TokenPair {
  user?: User
}

export async function getBaseUrl(): Promise<string> {
  return (await AsyncStorage.getItem(BASE_URL_KEY)) ?? DEFAULT_URL
}

export async function setServerUrl(url: string): Promise<void> {
  await AsyncStorage.setItem(BASE_URL_KEY, url.trim().replace(/\/+$/, ''))
}

export const getAccessToken = () => SecureStore.getItemAsync(ACCESS_TOKEN_KEY)
export const getRefreshToken = () => SecureStore.getItemAsync(REFRESH_TOKEN_KEY)

export async function setTokens(access: string, refresh?: string): Promise<void> {
  await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, access)
  if (refresh !== undefined) await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refresh)
}

export async function clearTokens(): Promise<void> {
  await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY)
  await SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY)
}

let authFailureHandler: (() => void) | null = null

/** Registered by the auth context; called after a failed refresh. Returns an unregister function. */
export function onAuthFailure(handler: () => void): () => void {
  authFailureHandler = handler
  return () => {
    if (authFailureHandler === handler) authFailureHandler = null
  }
}

export const api = axios.create()

api.interceptors.request.use(async (config) => {
  config.baseURL = await getBaseUrl()
  const token = await getAccessToken()
  if (token) config.headers.set('Authorization', `Bearer ${token}`)
  return config
})

let inFlightRefresh: Promise<RefreshResponse> | null = null

/** Exchanges the stored refresh token for new tokens. Concurrent callers share one request. */
export function refreshTokens(): Promise<RefreshResponse> {
  inFlightRefresh ??= doRefresh().finally(() => {
    inFlightRefresh = null
  })
  return inFlightRefresh
}

async function doRefresh(): Promise<RefreshResponse> {
  const refreshToken = await getRefreshToken()
  if (!refreshToken) throw new Error('no refresh token')
  const res = await api.post<RefreshResponse>(
    REFRESH_PATH,
    { refresh_token: refreshToken },
    { skipAuthRefresh: true },
  )
  await setTokens(res.data.access_token, res.data.refresh_token)
  return res.data
}

api.interceptors.response.use(undefined, async (error: unknown) => {
  if (!(error instanceof AxiosError) || error.response?.status !== 401 || !error.config) {
    throw error
  }
  const original = error.config
  if (original.authRetried || original.skipAuthRefresh) throw error
  original.authRetried = true
  try {
    const { access_token } = await refreshTokens()
    original.headers.set('Authorization', `Bearer ${access_token}`)
  } catch {
    await clearTokens()
    authFailureHandler?.()
    throw error
  }
  return api.request(original)
})
