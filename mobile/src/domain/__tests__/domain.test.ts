import { err, map, mapError, normalizeServerUrl, ok, rotateTokens } from '..'

describe('Result', () => {
  it('maps the value of an Ok and leaves an Err alone', () => {
    expect(map(ok(2), (n) => n * 2)).toEqual(ok(4))
    expect(map(err('boom'), (n: number) => n * 2)).toEqual(err('boom'))
  })

  it('maps the error of an Err and leaves an Ok alone', () => {
    expect(mapError(err('a'), (e) => `${e}!`)).toEqual(err('a!'))
    expect(mapError(ok(1), (e: string) => `${e}!`)).toEqual(ok(1))
  })
})

describe('rotateTokens', () => {
  it('takes the new access token and the new refresh token', () => {
    expect(rotateTokens({ accessToken: 'a1', refreshToken: 'r1' }, { accessToken: 'a2', refreshToken: 'r2' })).toEqual({
      accessToken: 'a2',
      refreshToken: 'r2',
    })
  })

  it('keeps the current refresh token when the server omits one', () => {
    expect(rotateTokens({ accessToken: 'a1', refreshToken: 'r1' }, { accessToken: 'a2' })).toEqual({
      accessToken: 'a2',
      refreshToken: 'r1',
    })
  })

  it('does not mutate its inputs', () => {
    const current = Object.freeze({ accessToken: 'a1', refreshToken: 'r1' })
    expect(() => rotateTokens(current, Object.freeze({ accessToken: 'a2' }))).not.toThrow()
  })
})

describe('normalizeServerUrl', () => {
  it.each([
    ['  https://x.example  ', 'https://x.example'],
    ['https://x.example/', 'https://x.example'],
    ['https://x.example///', 'https://x.example'],
    ['https://x.example/api', 'https://x.example/api'],
  ])('%j -> %j', (input, expected) => {
    expect(normalizeServerUrl(input)).toBe(expected)
  })
})
