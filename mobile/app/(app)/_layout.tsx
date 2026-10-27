import { Redirect, Stack } from 'expo-router'

import { useAuth } from '../../src/presentation/auth'
import { authRedirect } from '../../src/presentation/auth-guard'

export default function AppLayout() {
  const { user, isLoading } = useAuth()
  if (isLoading) return null
  const href = authRedirect({ group: 'app', signedIn: !!user, isLoading })
  if (href) return <Redirect href={href} />
  return <Stack screenOptions={{ headerShown: false }} />
}
