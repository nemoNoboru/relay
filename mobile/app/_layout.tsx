import { useFonts } from 'expo-font'
import { Slot } from 'expo-router'
import * as SplashScreen from 'expo-splash-screen'
import { StatusBar } from 'expo-status-bar'
import { useEffect } from 'react'

import { AppProviders } from '../src/composition/AppProviders'
import { fontSources } from '../src/presentation/fonts'

void SplashScreen.preventAutoHideAsync()

export default function RootLayout() {
  const [fontsLoaded, fontError] = useFonts(fontSources)
  const ready = fontsLoaded || !!fontError

  useEffect(() => {
    if (ready) void SplashScreen.hideAsync()
  }, [ready])

  if (!ready) return null
  return (
    <AppProviders>
      <StatusBar style="dark" />
      <Slot />
    </AppProviders>
  )
}
