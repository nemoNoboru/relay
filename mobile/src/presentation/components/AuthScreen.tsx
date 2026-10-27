import type { ReactNode } from 'react'
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, View } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'

import { colors } from '../theme'

type Props = { header: ReactNode; children: ReactNode; footer: ReactNode }

/** Poster screen frame for forms: header band on top, keyboard-avoiding scrolling form, links pinned below. */
export function AuthScreen({ header, children, footer }: Props) {
  const { bottom } = useSafeAreaInsets()
  return (
    <KeyboardAvoidingView style={styles.screen} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      {header}
      <ScrollView
        contentContainerStyle={[styles.content, { paddingBottom: Math.max(bottom, 16) + 16 }]}
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="on-drag"
      >
        <View>{children}</View>
        <View style={styles.footer}>{footer}</View>
      </ScrollView>
    </KeyboardAvoidingView>
  )
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.paper },
  content: { flexGrow: 1, justifyContent: 'space-between', paddingHorizontal: 24, paddingTop: 6 },
  footer: { alignItems: 'center', paddingTop: 32 },
})
