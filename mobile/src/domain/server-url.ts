export const DEFAULT_SERVER_URL = 'https://relay.mansus.org'

export const normalizeServerUrl = (url: string): string => url.trim().replace(/\/+$/, '')
