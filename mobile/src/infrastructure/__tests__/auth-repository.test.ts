/* eslint-disable import/no-named-as-default-member -- axios.create is the documented API */
import axios, { AxiosError } from 'axios'

import { err, ok } from '../../domain'
import { makeFakeServer, type FakeHandler } from '../../testing/fake-server'
import { makeHttpAuthRepository } from '../http/auth-repository'
import { LOGIN_PATH, REFRESH_PATH } from '../http/endpoints'

function setup(handler: FakeHandler) {
  const server = makeFakeServer(handler)
  const http = axios.create({ baseURL: 'https://x.example', adapter: server.adapter })
  return { repository: makeHttpAuthRepository({ http }), seen: server.seen }
}

describe('login', () => {
  it('posts the credentials and maps the response', async () => {
    const { repository, seen } = setup(() => ({
      status: 200,
      data: { access_token: 'a', refresh_token: 'r', user: { id: 'u1', name: 'Ada', extra: true } },
    }))
    const result = await repository.login({ username: 'ada', password: 'pw' })
    expect(result).toEqual(ok({ user: { id: 'u1', name: 'Ada' }, tokens: { accessToken: 'a', refreshToken: 'r' } }))
    expect(seen[0]!.url).toBe(LOGIN_PATH)
    expect(JSON.parse(seen[0]!.data)).toEqual({ username: 'ada', password: 'pw' })
  })

  it.each([[401], [403]])('maps %i to invalid-credentials', async (status) => {
    const { repository } = setup(() => ({ status }))
    expect(await repository.login({ username: 'a', password: 'b' })).toEqual(err('invalid-credentials'))
  })

  it('maps a server error to unexpected', async () => {
    const { repository } = setup(() => ({ status: 500 }))
    expect(await repository.login({ username: 'a', password: 'b' })).toEqual(err('unexpected'))
  })

  it('maps a missing response to network', async () => {
    const http = axios.create({
      adapter: async (config) => {
        throw new AxiosError('offline', 'ERR_NETWORK', config)
      },
    })
    const result = await makeHttpAuthRepository({ http }).login({ username: 'a', password: 'b' })
    expect(result).toEqual(err('network'))
  })
})

describe('refresh', () => {
  it('posts the refresh token; user and refresh token are optional in the response', async () => {
    const { repository, seen } = setup(() => ({ status: 200, data: { access_token: 'a2' } }))
    const result = await repository.refresh('r1')
    expect(result).toEqual(ok({ user: null, tokens: { accessToken: 'a2', refreshToken: undefined } }))
    expect(seen[0]!.url).toBe(REFRESH_PATH)
    expect(JSON.parse(seen[0]!.data)).toEqual({ refresh_token: 'r1' })
  })

  it('maps the user when present', async () => {
    const { repository } = setup(() => ({
      status: 200,
      data: { access_token: 'a2', refresh_token: 'r2', user: { id: 'u1', name: 'Ada' } },
    }))
    const result = await repository.refresh('r1')
    expect(result).toEqual(ok({ user: { id: 'u1', name: 'Ada' }, tokens: { accessToken: 'a2', refreshToken: 'r2' } }))
  })
})
