import { StyleSheet, Text, View } from 'react-native'

import { colors, fonts, line } from '../theme'
import { HardShadow } from './HardShadow'

/** Ink block with a red flag and a red offset: screen-level errors. */
export function Alert({ message }: { message: string }) {
  return (
    <HardShadow color={colors.red} style={styles.wrap}>
      <View style={styles.alert} accessibilityRole="alert" accessibilityLiveRegion="assertive">
        <View style={styles.flag} importantForAccessibility="no-hide-descendants" accessibilityElementsHidden>
          <Text style={styles.bang}>!</Text>
        </View>
        <Text style={styles.text}>{message}</Text>
      </View>
    </HardShadow>
  )
}

const styles = StyleSheet.create({
  wrap: { marginTop: 18 },
  alert: { flexDirection: 'row', alignItems: 'stretch', backgroundColor: colors.ink, borderWidth: line.strong, borderColor: colors.ink },
  flag: { width: 50, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.red },
  bang: { fontFamily: fonts.display, fontSize: 28, color: colors.onRed },
  text: { flex: 1, paddingVertical: 12, paddingHorizontal: 14, fontFamily: fonts.bodyMedium, fontSize: 16, lineHeight: 21, color: colors.paper },
})
