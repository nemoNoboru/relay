import type { ReactNode } from 'react'
import { StyleSheet, View, type StyleProp, type ViewStyle } from 'react-native'

import { colors } from '../theme'

type Props = {
  children: ReactNode
  /** Offset in px (5 for cards and buttons, 3 for small elements). */
  offset?: number
  color?: string
  /** Shifts the content onto its shadow, as a pressed button does. */
  pressed?: boolean
  style?: StyleProp<ViewStyle>
}

/** Hard offset shadow: an ink block behind the content, inside a wrapper that reserves the space. */
export function HardShadow({ children, offset = 5, color = colors.ink, pressed = false, style }: Props) {
  const shift = pressed ? offset : 0
  return (
    <View style={[{ paddingRight: offset, paddingBottom: offset }, style]}>
      <View pointerEvents="none" style={[styles.shadow, { top: offset, left: offset, backgroundColor: color }]} />
      <View style={{ transform: [{ translateX: shift }, { translateY: shift }] }}>{children}</View>
    </View>
  )
}

const styles = StyleSheet.create({
  shadow: { position: 'absolute', right: 0, bottom: 0 },
})
