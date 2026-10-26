import { createContext, useContext, type ReactNode } from 'react'

import type { UseCases } from '../application'

const UseCasesContext = createContext<UseCases | null>(null)

export function UseCasesProvider({ useCases, children }: { useCases: UseCases; children: ReactNode }) {
  return <UseCasesContext.Provider value={useCases}>{children}</UseCasesContext.Provider>
}

export function useUseCases(): UseCases {
  const useCases = useContext(UseCasesContext)
  if (!useCases) throw new Error('useUseCases must be used inside UseCasesProvider')
  return useCases
}
