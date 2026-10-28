import { Text, type StyleProp, type TextProps, type TextStyle } from 'react-native'

import { colors, fonts } from '../theme'

type Props = TextProps & { size?: number; tracking?: number; color?: string; style?: StyleProp<TextStyle> }

/** Space Mono, uppercase, tracked: labels, buttons, metadata. `tracking` is in em. */
export function MonoLabel({ children, size = 11, tracking = 0.16, color = colors.ink, style, ...rest }: Props) {
  return (
    <Text
      {...rest}
      style={[{ fontFamily: fonts.monoBold, fontSize: size, letterSpacing: size * tracking, color }, style]}
    >
      {typeof children === 'string' ? children.toUpperCase() : children}
    </Text>
  )
}
