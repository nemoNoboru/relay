import type { SessionEvents } from '../domain'

export type SessionEventBus = SessionEvents & { readonly notifyExpired: () => void }

export const makeSessionEvents = (): SessionEventBus => {
  const handlers = new Set<() => void>()
  return {
    onExpired(handler) {
      handlers.add(handler)
      return () => void handlers.delete(handler)
    },
    notifyExpired: () => handlers.forEach((handler) => handler()),
  }
}
