export interface RelayAPI {
  showOpenDialog: (options: Electron.OpenDialogOptions) => Promise<Electron.OpenDialogReturnValue>
  showSaveDialog: (options: Electron.SaveDialogOptions) => Promise<Electron.SaveDialogReturnValue>
  showMessageBox: (options: Electron.MessageBoxOptions) => Promise<Electron.MessageBoxReturnValue>
  watchProjectFiles: (projectPath: string) => Promise<boolean>
  onMenuNewProject: (callback: () => void) => void
  onMenuOpenProject: (callback: (projectPath: string) => void) => void
  onFileChanged: (callback: (filePath: string) => void) => void
  onFileAdded: (callback: (filePath: string) => void) => void
  onFileRemoved: (callback: (filePath: string) => void) => void
  removeAllListeners: (channel: string) => void
}

declare global {
  interface Window {
    relayAPI: RelayAPI
    ipcRenderer: {
      on: (channel: string, listener: (event: any, ...args: any[]) => void) => void
      off: (channel: string, ...args: any[]) => void
      send: (channel: string, ...args: any[]) => void
      invoke: (channel: string, ...args: any[]) => Promise<any>
    }
  }
} 