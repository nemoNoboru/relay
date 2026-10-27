import { Link, type Href } from 'expo-router'
import { StyleSheet } from 'react-native'

import { colors, fonts, minTap } from '../theme'

type Props = { href: Href; label: string; tone?: 'red' | 'ink'; replace?: boolean }

/** Underlined mono link (44px tall). Red for the main alternative, ink for quiet ones. */
export function LinkButton({ href, label, tone = 'red', replace }: Props) {
  return (
    <Link href={href} replace={replace} accessibilityRole="link" style={[styles.link, { color: tone === 'red' ? colors.redText : colors.ink }]}>
      {label.toUpperCase()}
    </Link>
  )
}

const styles = StyleSheet.create({
  link: {
    minHeight: minTap,
    paddingHorizontal: 12,
    paddingVertical: 13,
    fontFamily: fonts.monoBold,
    fontSize: 12.5,
    letterSpacing: 12.5 * 0.14,
    textAlign: 'center',
    textDecorationLine: 'underline',
  },
})
