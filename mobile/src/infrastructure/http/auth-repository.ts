import { AxiosError, type AxiosInstance } from 'axios'

import { err, ok, type AuthError, type AuthRepository } from '../../domain'
import { LOGIN_PATH, REFRESH_PATH } from './endpoints'

type UserDto = { id: string; name: string }
type LoginDto = { access_token: string; refresh_token: string; user: UserDto }
type RefreshDto = { access_token: string; refresh_token?: string; user?: UserDto }

const toAuthError = (error: unknown): AuthError => {
  if (!(error instanceof AxiosError)) return 'unexpected'
  if (!error.response) return 'network'
  return [400, 401, 403].includes(error.response.status) ? 'invalid-credentials' : 'unexpected'
}

const toUser = ({ id, name }: UserDto) => ({ id, name })

export const makeHttpAuthRepository = ({ http }: { http: AxiosInstance }): AuthRepository => ({
  async login(credentials) {
    try {
      const { data } = await http.post<LoginDto>(LOGIN_PATH, credentials, { skipAuthRefresh: true })
      return ok({
        user: toUser(data.user),
        tokens: { accessToken: data.access_token, refreshToken: data.refresh_token },
      })
    } catch (error) {
      return err(toAuthError(error))
    }
  },

  async refresh(refreshToken) {
    try {
      const { data } = await http.post<RefreshDto>(
        REFRESH_PATH,
        { refresh_token: refreshToken },
        { skipAuthRefresh: true },
      )
      return ok({
        user: data.user ? toUser(data.user) : null,
        tokens: { accessToken: data.access_token, refreshToken: data.refresh_token },
      })
    } catch (error) {
      return err(toAuthError(error))
    }
  },
})
