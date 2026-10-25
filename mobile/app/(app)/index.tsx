import { Pressable, StyleSheet, Text, View } from 'react-native'

import { useAuth } from '../../src/auth'
import { colors } from '../../src/theme'

export default function Home() {
  const { user, logout } = useAuth()
  return (
    <View style={styles.screen}>
      <Text style={styles.title}>RELAY</Text>
      <Text style={styles.body}>Signed in as {user?.name}</Text>
      <Pressable style={styles.button} onPress={() => void logout()}>
        <Text style={styles.buttonText}>Log out</Text>
      </Pressable>
    </View>
  )
}

const styles = StyleSheet.create({
  screen: { flex: 1, justifyContent: 'center', padding: 24, backgroundColor: colors.paper },
  title: { fontSize: 40, fontWeight: '900', color: colors.ink },
  body: { fontSize: 16, color: colors.ink2, marginVertical: 12 },
  button: { backgroundColor: colors.ink, padding: 14, alignItems: 'center' },
  buttonText: { color: colors.surface, fontWeight: '700' },
})
