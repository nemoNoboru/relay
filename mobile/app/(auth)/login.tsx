import { useState } from 'react'
import { Pressable, StyleSheet, Text, TextInput, View } from 'react-native'

import { DEFAULT_URL } from '../../src/api'
import { useAuth } from '../../src/auth'
import { colors } from '../../src/theme'

export default function Login() {
  const { login } = useAuth()
  const [serverUrl, setServerUrl] = useState(DEFAULT_URL)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function submit() {
    setBusy(true)
    setError(null)
    try {
      await login({ serverUrl, username, password })
    } catch {
      setError('Could not sign in. Check the server URL and your credentials.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <View style={styles.screen}>
      <Text style={styles.title}>RELAY</Text>
      <TextInput
        style={styles.input}
        value={serverUrl}
        onChangeText={setServerUrl}
        placeholder="Server URL"
        autoCapitalize="none"
        autoCorrect={false}
        keyboardType="url"
      />
      <TextInput
        style={styles.input}
        value={username}
        onChangeText={setUsername}
        placeholder="Username"
        autoCapitalize="none"
        autoCorrect={false}
      />
      <TextInput
        style={styles.input}
        value={password}
        onChangeText={setPassword}
        placeholder="Password or token"
        secureTextEntry
        autoCapitalize="none"
      />
      {error && <Text style={styles.error}>{error}</Text>}
      <Pressable style={styles.button} onPress={() => void submit()} disabled={busy}>
        <Text style={styles.buttonText}>{busy ? 'Signing in…' : 'Sign in'}</Text>
      </Pressable>
    </View>
  )
}

const styles = StyleSheet.create({
  screen: { flex: 1, justifyContent: 'center', padding: 24, gap: 12, backgroundColor: colors.paper },
  title: { fontSize: 40, fontWeight: '900', color: colors.ink, marginBottom: 12 },
  input: { borderWidth: 2, borderColor: colors.ink, backgroundColor: colors.surface, padding: 12, fontSize: 16 },
  error: { color: colors.red },
  button: { backgroundColor: colors.ink, padding: 14, alignItems: 'center' },
  buttonText: { color: colors.surface, fontWeight: '700' },
})
