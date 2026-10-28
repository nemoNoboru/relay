import { useState } from 'react'
import { StyleSheet, Text, TextInput, View, type TextInputProps } from 'react-native'

import { colors, fonts, line, minTap } from '../theme'
import { HardShadow } from './HardShadow'
import { MonoLabel } from './MonoLabel'

type Props = Omit<TextInputProps, 'style'> & {
  label: string
  error?: string | null
  /** Marks the field red without a message (a message elsewhere explains it). */
  invalid?: boolean
}

/** Labelled input with a red rule before the label. Focus and error use a 3px border and a red offset. */
export function TextField({ label, error, invalid = false, onFocus, onBlur, ...input }: Props) {
  const [focused, setFocused] = useState(false)
  const flagged = invalid || !!error
  const emphasised = flagged || focused
  return (
    <View style={styles.field}>
      <View style={styles.labelRow}>
        <View style={styles.rule} />
        <MonoLabel>{label}</MonoLabel>
      </View>
      <HardShadow offset={emphasised ? 4 : 0} color={colors.red}>
        <TextInput
          accessibilityLabel={label}
          placeholderTextColor={colors.ink2}
          selectionColor={colors.red}
          {...input}
          onFocus={(e) => {
            setFocused(true)
            onFocus?.(e)
          }}
          onBlur={(e) => {
            setFocused(false)
            onBlur?.(e)
          }}
          style={[
            styles.input,
            emphasised && { borderWidth: line.strong, paddingHorizontal: 13, paddingVertical: 11 },
            flagged && { borderColor: colors.redText },
          ]}
        />
      </HardShadow>
      {error ? (
        <View style={styles.errorRow} accessibilityRole="alert" accessibilityLiveRegion="polite">
          <Text style={styles.bang}>!</Text>
          <MonoLabel size={12} tracking={0.08} color={colors.redText}>
            {error}
          </MonoLabel>
        </View>
      ) : null}
    </View>
  )
}

const styles = StyleSheet.create({
  field: { marginTop: 18, gap: 8 },
  labelRow: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  rule: { width: 14, height: 3, backgroundColor: colors.red },
  input: {
    minHeight: Math.max(50, minTap),
    paddingHorizontal: 14,
    paddingVertical: 12,
    borderWidth: line.regular,
    borderColor: colors.ink,
    backgroundColor: colors.surface,
    color: colors.ink,
    fontFamily: fonts.body,
    fontSize: 17,
  },
  errorRow: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  bang: { fontFamily: fonts.display, fontSize: 18, color: colors.redText },
})
