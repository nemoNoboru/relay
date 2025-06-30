import { contextBridge, ipcRenderer } from 'electron'

// --------- Expose some API to the Renderer process ---------
contextBridge.exposeInMainWorld('ipcRenderer', {
  on(...args: Parameters<typeof ipcRenderer.on>) {
    const [channel, listener] = args
    return ipcRenderer.on(channel, (event, ...args) => listener(event, ...args))
  },
  off(...args: Parameters<typeof ipcRenderer.off>) {
    const [channel, ...omit] = args
    return ipcRenderer.off(channel, ...omit)
  },
  send(...args: Parameters<typeof ipcRenderer.send>) {
    const [channel, ...omit] = args
    return ipcRenderer.send(channel, ...omit)
  },
  invoke(...args: Parameters<typeof ipcRenderer.invoke>) {
    const [channel, ...omit] = args
    return ipcRenderer.invoke(channel, ...omit)
  },
})

// Expose Relay-specific APIs
contextBridge.exposeInMainWorld('relayAPI', {
  // File system operations
  showOpenDialog: (options: Electron.OpenDialogOptions) =>
    ipcRenderer.invoke('show-open-dialog', options),
  
  showSaveDialog: (options: Electron.SaveDialogOptions) =>
    ipcRenderer.invoke('show-save-dialog', options),
  
  showMessageBox: (options: Electron.MessageBoxOptions) =>
    ipcRenderer.invoke('show-message-box', options),

  // Project file watching
  watchProjectFiles: (projectPath: string) =>
    ipcRenderer.invoke('watch-project-files', projectPath),

  // Event listeners
  onMenuNewProject: (callback: () => void) =>
    ipcRenderer.on('menu-new-project', callback),
  
  onMenuOpenProject: (callback: (projectPath: string) => void) =>
    ipcRenderer.on('menu-open-project', (event, projectPath) => callback(projectPath)),
  
  onFileChanged: (callback: (filePath: string) => void) =>
    ipcRenderer.on('file-changed', (event, filePath) => callback(filePath)),
  
  onFileAdded: (callback: (filePath: string) => void) =>
    ipcRenderer.on('file-added', (event, filePath) => callback(filePath)),
  
  onFileRemoved: (callback: (filePath: string) => void) =>
    ipcRenderer.on('file-removed', (event, filePath) => callback(filePath)),

  // Clean up listeners
  removeAllListeners: (channel: string) =>
    ipcRenderer.removeAllListeners(channel)
})

// Global type definitions for TypeScript
declare global {
  interface Window {
    ipcRenderer: typeof ipcRenderer
    relayAPI: {
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
  }
} 