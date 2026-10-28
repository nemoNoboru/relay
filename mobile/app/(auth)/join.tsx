import { useLocalSearchParams } from 'expo-router'
import { useState } from 'react'
import { View } from 'react-native'

import type { AuthError } from '../../src/domain'
import { useAuth } from '../../src/presentation/auth'
import { Alert } from '../../src/presentation/components/Alert'
import { AuthScreen } from '../../src/presentation/components/AuthScreen'
import { Button } from '../../src/presentation/components/Button'
import { GeoBar, GeoRing, GeoSquare, GeoWedge } from '../../src/presentation/components/Geo'
import { LinkButton } from '../../src/presentation/components/LinkButton'
import { Hot, PosterHeader } from '../../src/presentation/components/PosterHeader'
import { TextField } from '../../src/presentation/components/TextField'
import { colors } from '../../src/presentation/theme'

const INVALID_INVITE = 'This invite link is invalid or has already been used. Ask for a new one.'

const alertMessages: Partial<Record<AuthError, string>> = {
  'invalid-invite': INVALID_INVITE,
  network: 'Could not reach the server. Check your connection and try again.',
  unexpected: 'Something went wrong. Try again.',
}

export default function Join() {
  const { token } = useLocalSearchParams<{ token?: string | string[] }>()
  const inviteToken = (Array.isArray(token) ? token[0] : token)?.trim() ?? ''
  const { register } = useAuth()
  const [displayName, setDisplayName] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<AuthError | null>(null)
  const [busy, setBusy] = useState(false)

  const inviteInvalid = inviteToken === '' || error === 'invalid-invite'
  const alert = inviteToken === '' ? INVALID_INVITE : error ? alertMessages[error] : undefined
  const canSubmit = !inviteInvalid && displayName.trim() !== '' && username.trim() !== '' && password !== ''

  async function submit() {
    if (!canSubmit || busy) return
    setBusy(true)
    setError(null)
    const result = await register({ inviteToken, displayName, username, password })
    // On success the auth guard moves the user to the feed as soon as the session is set.
    if (!result.ok) setError(result.error)
    setBusy(false)
  }

  return (
    <AuthScreen
      header={
        <PosterHeader
          accessibleTitle="You were invited"
          title={
            <>
              You were{'\n'}
              <Hot>invited</Hot>
            </>
          }
          meta="to relay.mansus.org"
          decoration={
            <>
              <GeoRing size={140} right={-40} top={-10} color={colors.ochre} />
              <GeoSquare size={36} right={30} top={80} />
              {inviteInvalid ? (
                <GeoBar width={180} right={-30} top={130} />
              ) : (
                <GeoWedge size={140} right={-30} top={130} color={colors.ochre} flip />
              )}
            </>
          }
        />
      }
      footer={<LinkButton href="/login" label="Already have an account? Sign in" tone="ink" replace />}
    >
      {alert ? <Alert message={alert} /> : null}
      <TextField
        label="Display name"
        value={displayName}
        onChangeText={setDisplayName}
        autoCapitalize="words"
        autoComplete="name"
        textContentType="name"
        returnKeyType="next"
      />
      <TextField
        label="Username"
        value={username}
        onChangeText={(value) => {
          setUsername(value)
          if (error === 'username-taken') setError(null)
        }}
        error={error === 'username-taken' ? 'That username is taken' : null}
        autoCapitalize="none"
        autoCorrect={false}
        autoComplete="username-new"
        textContentType="username"
        returnKeyType="next"
      />
      <TextField
        label="Password"
        value={password}
        onChangeText={setPassword}
        secureTextEntry
        autoCapitalize="none"
        autoComplete="new-password"
        textContentType="newPassword"
        returnKeyType="go"
        onSubmitEditing={() => void submit()}
      />
      <View style={{ marginTop: 26 }}>
        <Button label={busy ? 'Joining…' : 'Join'} onPress={() => void submit()} busy={busy} disabled={!canSubmit} />
      </View>
    </AuthScreen>
  )
}
