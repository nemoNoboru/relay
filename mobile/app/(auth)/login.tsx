import { useState } from 'react'
import { View } from 'react-native'

import type { AuthError } from '../../src/domain'
import { useAuth } from '../../src/presentation/auth'
import { Alert } from '../../src/presentation/components/Alert'
import { AuthScreen } from '../../src/presentation/components/AuthScreen'
import { Button } from '../../src/presentation/components/Button'
import { GeoRing, GeoSun, GeoWedge } from '../../src/presentation/components/Geo'
import { LinkButton } from '../../src/presentation/components/LinkButton'
import { Hot, PosterHeader } from '../../src/presentation/components/PosterHeader'
import { TextField } from '../../src/presentation/components/TextField'

const messages: Partial<Record<AuthError, string>> = {
  'invalid-credentials': 'Wrong username or password.',
  network: 'Could not reach the server. Check your connection and try again.',
}

export default function Login() {
  const { login } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<AuthError | null>(null)
  const [busy, setBusy] = useState(false)

  const canSubmit = username.trim() !== '' && password !== ''

  async function submit() {
    if (!canSubmit || busy) return
    setBusy(true)
    setError(null)
    // A blank server URL keeps the stored one; changing server lives in settings.
    const result = await login({ serverUrl: '', username: username.trim(), password })
    // On success the auth guard moves the user to the feed as soon as the session is set.
    if (!result.ok) setError(result.error)
    setBusy(false)
  }

  return (
    <AuthScreen
      header={
        <PosterHeader
          accessibleTitle="Sign in to your node"
          title={
            <>
              Sign in{'\n'}to your <Hot>node</Hot>
            </>
          }
          meta="relay.mansus.org"
          decoration={
            <>
              <GeoSun size={120} right={-34} top={20} />
              <GeoRing size={160} right={-60} top={0} />
              <GeoWedge size={120} right={-24} top={124} />
            </>
          }
        />
      }
      footer={<LinkButton href="/join" label="Join with invite link" />}
    >
      {error ? <Alert message={messages[error] ?? 'Something went wrong. Try again.'} /> : null}
      <TextField
        label="Username"
        value={username}
        onChangeText={setUsername}
        invalid={error === 'invalid-credentials'}
        autoCapitalize="none"
        autoCorrect={false}
        autoComplete="username"
        textContentType="username"
        returnKeyType="next"
      />
      <TextField
        label="Password"
        value={password}
        onChangeText={setPassword}
        invalid={error === 'invalid-credentials'}
        secureTextEntry
        autoCapitalize="none"
        autoComplete="current-password"
        textContentType="password"
        returnKeyType="go"
        onSubmitEditing={() => void submit()}
      />
      <View style={{ marginTop: 26 }}>
        <Button label={busy ? 'Signing in…' : 'Sign in'} onPress={() => void submit()} busy={busy} disabled={!canSubmit} />
      </View>
    </AuthScreen>
  )
}
