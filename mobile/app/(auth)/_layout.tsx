import { Redirect, Stack } from 'expo-router'
import { StatusBar } from 'expo-status-bar'

import { useAuth } from '../../src/presentation/auth'
import { authRedirect } from '../../src/presentation/auth-guard'

export default function AuthLayout() {
  const { user, isLoading } = useAuth()
  if (isLoading) return null
  const href = authRedirect({ group: 'auth', signedIn: !!user, isLoading })
  if (href) return <Redirect href={href} />
  return (
    <>
      <StatusBar style="light" />
      <Stack screenOptions={{ headerShown: false }} />
    </>
  )
}
