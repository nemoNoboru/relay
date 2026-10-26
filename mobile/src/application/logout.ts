import type { TokenWriter } from '../domain'

type Deps = { readonly tokenWriter: Pick<TokenWriter, 'clear'> }

export const makeLogout =
  ({ tokenWriter }: Deps) =>
  (): Promise<void> =>
    tokenWriter.clear()
