// Tokens from mobile-mockups/shared.css (see .pi/skills/relay-mobile-design).
export const colors = {
  paper: '#f0e6d2',
  paper2: '#e4d5b7',
  surface: '#f8f1e3',
  ink: '#141210',
  ink2: '#5c544a',
  red: '#ce1f2b',
  redText: '#b81a25',
  redDeep: '#a5141f',
  ochre: '#dfa32a',
  onRed: '#fff8ea',
  hair: 'rgba(20,18,16,.22)',
} as const

export const line = { regular: 2, strong: 3 } as const

export const shadow = { regular: 5, small: 3 } as const

/** Exact family names registered by `useFonts` in app/_layout.tsx. */
export const fonts = {
  display: 'Anton_400Regular',
  body: 'Oswald_400Regular',
  bodyMedium: 'Oswald_500Medium',
  mono: 'SpaceMono_400Regular',
  monoBold: 'SpaceMono_700Bold',
} as const

export const minTap = 44
