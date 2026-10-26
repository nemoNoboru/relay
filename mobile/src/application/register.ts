import { err, ok, type AuthError, type AuthRepository, type Registration, type Result, type TokenWriter, type User } from '../domain'

export type RegisterInput = Registration

type Deps = {
  readonly authRepository: Pick<AuthRepository, 'register'>
  readonly tokenWriter: Pick<TokenWriter, 'save'>
}

/** Redeems an invite token for a new account and signs the user in. */
export const makeRegister =
  ({ authRepository, tokenWriter }: Deps) =>
  async (input: RegisterInput): Promise<Result<User, AuthError>> => {
    const result = await authRepository.register({
      ...input,
      displayName: input.displayName.trim(),
      username: input.username.trim(),
    })
    if (!result.ok) return err(result.error)
    await tokenWriter.save(result.value.tokens)
    return ok(result.value.user)
  }
