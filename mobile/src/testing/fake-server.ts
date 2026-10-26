import { AxiosError, type AxiosAdapter, type InternalAxiosRequestConfig } from 'axios'

export type FakeResponse = { status: number; data?: unknown }
export type FakeHandler = (config: InternalAxiosRequestConfig) => FakeResponse

/** An axios adapter that answers from a handler instead of the network and records every request. */
export function makeFakeServer(handler: FakeHandler) {
  const seen: InternalAxiosRequestConfig[] = []
  const adapter: AxiosAdapter = async (config) => {
    seen.push(config)
    const { status, data = {} } = handler(config)
    const response = { data, status, statusText: String(status), headers: {}, config }
    if (status >= 400) throw new AxiosError('failed', 'ERR_BAD_REQUEST', config, null, response)
    return response
  }
  return { adapter, seen }
}

export const urlOf = (config: InternalAxiosRequestConfig) => `${config.baseURL}${config.url}`
