import { useState } from 'react'
import { Pressable, StyleSheet } from 'react-native'

import { colors, line, minTap } from '../theme'
import { HardShadow } from './HardShadow'
import { MonoLabel } from './MonoLabel'

type Props = {
  label: string
  onPress: () => void
  disabled?: boolean
  busy?: boolean
}

/** Red block button with a hard ink shadow. Disabled turns paper2 with muted text and no shadow. */
export function Button({ label, onPress, disabled = false, busy = false }: Props) {
  const [pressed, setPressed] = useState(false)
  const inactive = disabled || busy
  const face = (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityState={{ disabled: inactive, busy }}
      disabled={inactive}
      onPress={onPress}
      onPressIn={() => setPressed(true)}
      onPressOut={() => setPressed(false)}
      style={[styles.button, inactive ? styles.inactive : styles.active]}
    >
      <MonoLabel size={14} tracking={0.16} color={inactive ? colors.ink2 : colors.onRed}>
        {label}
      </MonoLabel>
    </Pressable>
  )
  if (inactive) return <HardShadow offset={0}>{face}</HardShadow>
  return <HardShadow pressed={pressed}>{face}</HardShadow>
}

const styles = StyleSheet.create({
  button: { height: 50, minHeight: minTap, borderWidth: line.strong, alignItems: 'center', justifyContent: 'center' },
  active: { backgroundColor: colors.red, borderColor: colors.ink },
  inactive: { backgroundColor: colors.paper2, borderColor: colors.ink2 },
})
