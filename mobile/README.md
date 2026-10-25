# Relay mobile

Expo (SDK 58) / React Native app, TypeScript strict, Expo Router. Design: `.pi/skills/relay-mobile-design`.

## Run

```sh
cd mobile
npm install
npx expo start      # then press i (iOS simulator) or a (Android emulator)
```

The server URL defaults to `https://relay.mansus.org` and can be changed on the login screen
(stored in AsyncStorage as `relay_server_url`, applied on the next request).

## Layout

- `app/` file-based routes: `_layout.tsx` (AuthProvider), `(auth)/login`, `(app)/index`
- `src/api.ts` axios client: per-request base URL, Bearer token, single-flight 401 refresh
- `src/auth.tsx` `AuthProvider` / `useAuth()`
- Tokens live in SecureStore (`access_token`, `refresh_token`). Auth endpoint paths are constants in `src/api.ts`; the backend does not serve them yet.

## Checks

```sh
npm test
npm run typecheck
npx expo-doctor
npx expo export --platform ios --platform android   # bundle check; dist/ is ignored
```
