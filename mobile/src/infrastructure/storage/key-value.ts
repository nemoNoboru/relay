export type KeyValueStorage = {
  readonly get: (key: string) => Promise<string | null>
  readonly set: (key: string, value: string) => Promise<void>
  readonly remove: (key: string) => Promise<void>
}
